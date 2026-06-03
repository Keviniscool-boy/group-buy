package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"zhixiang-group-buying/models"
)

// ─── 微信端订单 ──────────────────────────────────

// WxCreateOrder POST /wx/orders
func WxCreateOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		var req struct {
			Items    []models.CartItem `json:"items"`
			StoreID  uint              `json:"store_id"`
			Remark   string            `json:"remark"`
			CouponID uint              `json:"coupon_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || len(req.Items) == 0 {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误或购物车为空"})
			return
		}

		// 获取用户资讯
		var user models.User
		db.First(&user, userID)

		// 计算总金额并检查库存
		var total float64
		for _, it := range req.Items {
			var com models.Commodity
			if db.First(&com, it.CommodityID).Error != nil {
				continue
			}
			// 预售检查：若未到销售开始时间，不允许购买
			if com.SaleStartTime != nil && com.SaleStartTime.After(time.Now()) {
				c.JSON(http.StatusOK, models.Result{Code: 400, Msg: fmt.Sprintf("商品「%s」尚未开售", com.Name)})
				return
			}
			if com.Status == 0 {
				c.JSON(http.StatusOK, models.Result{Code: 400, Msg: fmt.Sprintf("商品「%s」已下架", com.Name)})
				return
			}
			// 团购价
			price := com.Price
			if com.IsGroupon == 1 && com.GroupPrice > 0 && com.GroupStartTime != nil && com.GroupStartTime.Before(time.Now()) {
				price = com.GroupPrice
			}
			// 使用请求中的数量
			qty := it.Quantity
			if qty <= 0 {
				qty = 1
			}
			if com.Stock < qty {
				c.JSON(http.StatusOK, models.Result{Code: 400, Msg: fmt.Sprintf("商品「%s」库存不足", com.Name)})
				return
			}
			total += price * float64(qty)
		}

		var selectedCoupon models.UserCoupon
		discount := 0.0
		payAmount := total
		if req.CouponID > 0 {
			if db.Where("id = ? AND user_id = ? AND status = 0", req.CouponID, userID).First(&selectedCoupon).Error != nil {
				c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "优惠券不可用"})
				return
			}
			var coupon models.Coupon
			if db.First(&coupon, selectedCoupon.CouponID).Error != nil || !couponInDate(coupon, time.Now()) {
				c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "优惠券已失效"})
				return
			}
			discount = applyCouponDiscount(total, selectedCoupon)
			if discount <= 0 {
				c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "订单金额未达到优惠券使用门槛"})
				return
			}
			payAmount = total - discount
		}

		// 生成订单编号
		orderNo := fmt.Sprintf("ZX%d%04d", time.Now().Unix(), rand.Intn(10000))

		// 取出门店
		var storeName string
		if req.StoreID > 0 {
			var store models.Store
			if db.First(&store, req.StoreID).Error == nil {
				storeName = store.Name
			}
		}

		order := models.Order{
			OrderNo:        orderNo,
			UserID:         userID,
			UserName:       user.Nickname,
			TotalAmount:    total,
			DiscountAmount: discount,
			PayAmount:      payAmount,
			CouponID:       selectedCoupon.ID,
			CouponName:     selectedCoupon.CouponName,
			Status:         1, // 已付款（模拟）
			StoreID:        req.StoreID,
			StoreName:      storeName,
			PickupCode:     fmt.Sprintf("%04d", rand.Intn(10000)),
			Remark:         req.Remark,
		}

		tx := db.Begin()
		if err := tx.Create(&order).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusOK, models.Result{Code: 500, Msg: "建立订单失败"})
			return
		}

		if selectedCoupon.ID > 0 {
			now := time.Now()
			if err := tx.Model(&models.UserCoupon{}).Where("id = ? AND status = 0", selectedCoupon.ID).Updates(map[string]interface{}{
				"status":   1,
				"order_id": order.ID,
				"used_at":  &now,
			}).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusOK, models.Result{Code: 500, Msg: "优惠券使用失败"})
				return
			}
			tx.Model(&models.Coupon{}).Where("id = ?", selectedCoupon.CouponID).Update("used", gorm.Expr("used + 1"))
		}

		for _, it := range req.Items {
			var com models.Commodity
			if tx.First(&com, it.CommodityID).Error != nil {
				continue
			}
			price := com.Price
			if com.IsGroupon == 1 && com.GroupPrice > 0 && com.GroupStartTime != nil && com.GroupStartTime.Before(time.Now()) {
				price = com.GroupPrice
			}
			qty := it.Quantity
			if qty <= 0 {
				qty = 1
			}
			tx.Create(&models.OrderItem{
				OrderID:       order.ID,
				CommodityID:   com.ID,
				CommodityName: com.Name,
				Price:         price,
				Quantity:      qty,
				Image:         com.Image,
			})
			// 更新销量与库存（防止并发超卖）
			res := tx.Model(&models.Commodity{}).
				Where("id = ? AND stock >= ?", com.ID, qty).
				Updates(map[string]interface{}{
					"sales": gorm.Expr("sales + ?", qty),
					"stock": gorm.Expr("stock - ?", qty),
				})
			if res.Error != nil || res.RowsAffected == 0 {
				tx.Rollback()
				c.JSON(http.StatusOK, models.Result{Code: 400, Msg: fmt.Sprintf("商品「%s」库存不足", com.Name)})
				return
			}
			CreateStockLog(tx, com, -qty, com.Stock, com.Stock-qty, "order", order.ID, "订单扣减库存", "wx")
		}

		// 清空购物车中已结算的项目
		for _, it := range req.Items {
			tx.Delete(&models.CartItem{}, it.ID)
		}

		tx.Commit()

		// 发送订单消息
		db.Create(&models.Message{
			UserID:  userID,
			Title:   "订单建立成功",
			Content: fmt.Sprintf("您的订单 %s 已建立，实付 ¥%.2f，取货码 %s", orderNo, payAmount, order.PickupCode),
			Type:    1,
		})

		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "下单成功", Data: order})
	}
}

// WxListOrders GET /wx/orders
func WxListOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var list []models.Order
		db.Where("user_id = ?", userID).Order("id desc").Find(&list)

		// 为每个订单加载订单项
		type OrderWithItems struct {
			models.Order
			Items []models.OrderItem `json:"items"`
		}
		var result []OrderWithItems
		for _, o := range list {
			var items []models.OrderItem
			db.Where("order_id = ?", o.ID).Find(&items)
			result = append(result, OrderWithItems{Order: o, Items: items})
		}

		c.JSON(http.StatusOK, models.Result{Code: 0, Data: result})
	}
}

// WxConfirmOrder PUT /wx/orders/:id/confirm
func WxConfirmOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		var order models.Order
		if db.Where("id = ? AND user_id = ?", id, userID).First(&order).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "订单不存在"})
			return
		}
		if order.Status != 2 {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "订单状态不正确"})
			return
		}
		order.Status = 3 // 已取货
		db.Save(&order)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "确认收货成功"})
	}
}

// ─── 后台订单管理 ──────────────────────────────────

// AdminListOrders GET /admin/orders
func AdminListOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		keyword := c.Query("keyword")
		status := c.Query("status")

		query := db.Model(&models.Order{})
		if keyword != "" {
			query = query.Where("order_no LIKE ? OR user_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}

		var total int64
		query.Count(&total)

		var list []models.Order
		query.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&list)

		c.JSON(http.StatusOK, models.PageResult{Code: 0, Msg: "ok", Count: total, Data: list})
	}
}

// AdminGetOrder GET /admin/orders/:id
func AdminGetOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var order models.Order
		if db.First(&order, id).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "订单不存在"})
			return
		}
		var items []models.OrderItem
		db.Where("order_id = ?", order.ID).Find(&items)
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: gin.H{"order": order, "items": items}})
	}
}

// AdminUpdateOrderStatus PUT /admin/orders/:id/status
func AdminUpdateOrderStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Status  int  `json:"status"`
			StoreID uint `json:"store_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}

		var order models.Order
		if db.First(&order, id).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "订单不存在"})
			return
		}

		oldStatus := order.Status
		tx := db.Begin()

		order.Status = req.Status
		if req.StoreID > 0 {
			order.StoreID = req.StoreID
			var store models.Store
			if tx.First(&store, req.StoreID).Error == nil {
				order.StoreName = store.Name
			}
		}
		if order.Status == 2 && order.PickupCode == "" {
			order.PickupCode = fmt.Sprintf("%04d", rand.Intn(10000))
		}
		if order.Status == 3 && order.VerifyTime == nil {
			now := time.Now()
			order.VerifyBy = c.GetString("username")
			if order.VerifyBy == "" {
				order.VerifyBy = "admin"
			}
			order.VerifyTime = &now
		}

		if oldStatus != 4 && order.Status == 4 {
			var items []models.OrderItem
			tx.Where("order_id = ?", order.ID).Find(&items)
			for _, item := range items {
				var com models.Commodity
				if tx.First(&com, item.CommodityID).Error != nil {
					continue
				}
				before := com.Stock
				after := before + item.Quantity
				if err := tx.Model(&models.Commodity{}).Where("id = ?", com.ID).Updates(map[string]interface{}{
					"stock": gorm.Expr("stock + ?", item.Quantity),
					"sales": gorm.Expr("CASE WHEN sales >= ? THEN sales - ? ELSE 0 END", item.Quantity, item.Quantity),
				}).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusOK, models.Result{Code: 500, Msg: "库存回滚失败"})
					return
				}
				CreateStockLog(tx, com, item.Quantity, before, after, "cancel", order.ID, "订单取消回滚库存", c.GetString("username"))
			}

			if order.CouponID > 0 {
				var userCoupon models.UserCoupon
				tx.First(&userCoupon, order.CouponID)
				tx.Model(&models.UserCoupon{}).Where("id = ? AND order_id = ?", order.CouponID, order.ID).Updates(map[string]interface{}{
					"status":   0,
					"order_id": 0,
					"used_at":  nil,
				})
				if userCoupon.CouponID > 0 {
					tx.Model(&models.Coupon{}).Where("id = ? AND used > 0", userCoupon.CouponID).Update("used", gorm.Expr("used - 1"))
				}
			}
		}

		if err := tx.Save(&order).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusOK, models.Result{Code: 500, Msg: "更新失败"})
			return
		}

		// 状态变更通知
		statusText := map[int]string{0: "待付款", 1: "已付款", 2: "待取货", 3: "已取货", 4: "已取消"}
		tx.Create(&models.Message{
			UserID:  order.UserID,
			Title:   "订单状态更新",
			Content: fmt.Sprintf("订单 %s 状态已更新为「%s」", order.OrderNo, statusText[order.Status]),
			Type:    1,
		})
		tx.Commit()

		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "更新成功"})
	}
}

// AdminListOrderItems GET /admin/orders/:id/items
func AdminListOrderItems(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var items []models.OrderItem
		db.Where("order_id = ?", id).Find(&items)
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: items})
	}
}
