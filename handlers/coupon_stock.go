package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"zhixiang-group-buying/models"
)

func couponUsable(coupon models.Coupon, now time.Time) bool {
	if coupon.Status != 1 {
		return false
	}
	if coupon.StartTime != nil && coupon.StartTime.After(now) {
		return false
	}
	if coupon.EndTime != nil && coupon.EndTime.Before(now) {
		return false
	}
	return coupon.Total <= 0 || coupon.Received < coupon.Total
}

func couponInDate(coupon models.Coupon, now time.Time) bool {
	if coupon.Status != 1 {
		return false
	}
	if coupon.StartTime != nil && coupon.StartTime.After(now) {
		return false
	}
	return coupon.EndTime == nil || !coupon.EndTime.Before(now)
}

// CreateStockLog writes a single inventory movement record.
func CreateStockLog(tx *gorm.DB, com models.Commodity, change int, before int, after int, typ string, refID uint, remark string, operator string) {
	if operator == "" {
		operator = "system"
	}
	tx.Create(&models.StockLog{
		CommodityID:   com.ID,
		CommodityName: com.Name,
		ChangeQty:     change,
		BeforeQty:     before,
		AfterQty:      after,
		Type:          typ,
		RefID:         refID,
		Remark:        remark,
		Operator:      operator,
	})
}

// AdminListCoupons GET /admin/coupons
func AdminListCoupons(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		keyword := strings.TrimSpace(c.Query("keyword"))
		status := c.Query("status")

		query := db.Model(&models.Coupon{})
		if keyword != "" {
			query = query.Where("name LIKE ?", "%"+keyword+"%")
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}

		var total int64
		query.Count(&total)

		var list []models.Coupon
		query.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&list)
		c.JSON(http.StatusOK, models.PageResult{Code: 0, Msg: "ok", Count: total, Data: list})
	}
}

// CreateCoupon POST /admin/coupons
func CreateCoupon(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var coupon models.Coupon
		if err := c.ShouldBindJSON(&coupon); err != nil || coupon.Name == "" || coupon.Amount <= 0 {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "优惠券参数不正确"})
			return
		}
		db.Create(&coupon)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "优惠券创建成功", Data: coupon})
	}
}

// UpdateCoupon PUT /admin/coupons/:id
func UpdateCoupon(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var coupon models.Coupon
		if db.First(&coupon, id).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "优惠券不存在"})
			return
		}
		if err := c.ShouldBindJSON(&coupon); err != nil || coupon.Name == "" || coupon.Amount <= 0 {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "优惠券参数不正确"})
			return
		}
		db.Save(&coupon)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "优惠券更新成功", Data: coupon})
	}
}

// DeleteCoupon DELETE /admin/coupons/:id
func DeleteCoupon(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db.Delete(&models.Coupon{}, c.Param("id"))
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "优惠券删除成功"})
	}
}

// WxListCoupons GET /wx/coupons
func WxListCoupons(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.Coupon
		db.Where("status = 1").Order("id desc").Find(&list)
		now := time.Now()
		usable := make([]models.Coupon, 0, len(list))
		for _, coupon := range list {
			if couponUsable(coupon, now) {
				usable = append(usable, coupon)
			}
		}
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: usable})
	}
}

// WxReceiveCoupon POST /wx/coupons/:id/receive
func WxReceiveCoupon(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var coupon models.Coupon
		if db.First(&coupon, c.Param("id")).Error != nil || !couponUsable(coupon, time.Now()) {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "优惠券不可领取"})
			return
		}

		var exists int64
		db.Model(&models.UserCoupon{}).Where("user_id = ? AND coupon_id = ? AND status = 0", userID, coupon.ID).Count(&exists)
		if exists > 0 {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "你已经领取过这张优惠券"})
			return
		}

		tx := db.Begin()
		userCoupon := models.UserCoupon{
			UserID:     userID,
			CouponID:   coupon.ID,
			CouponName: coupon.Name,
			Threshold:  coupon.Threshold,
			Amount:     coupon.Amount,
			Status:     0,
		}
		if err := tx.Create(&userCoupon).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusOK, models.Result{Code: 500, Msg: "领取失败"})
			return
		}
		tx.Model(&models.Coupon{}).Where("id = ?", coupon.ID).Update("received", gorm.Expr("received + 1"))
		tx.Commit()

		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "领取成功", Data: userCoupon})
	}
}

// WxListMyCoupons GET /wx/my-coupons
func WxListMyCoupons(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		status := c.Query("status")
		total, _ := strconv.ParseFloat(c.DefaultQuery("total", "0"), 64)

		query := db.Where("user_id = ?", userID)
		if status != "" {
			query = query.Where("status = ?", status)
		}

		var list []models.UserCoupon
		query.Order("id desc").Find(&list)
		if total > 0 {
			filtered := make([]models.UserCoupon, 0, len(list))
			for _, coupon := range list {
				if coupon.Status == 0 && total >= coupon.Threshold {
					filtered = append(filtered, coupon)
				}
			}
			list = filtered
		}
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: list})
	}
}

// AdminListStockLogs GET /admin/stock-logs
func AdminListStockLogs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		keyword := strings.TrimSpace(c.Query("keyword"))
		typ := c.Query("type")

		query := db.Model(&models.StockLog{})
		if keyword != "" {
			query = query.Where("commodity_name LIKE ? OR remark LIKE ? OR operator LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}
		if typ != "" {
			query = query.Where("type = ?", typ)
		}

		var total int64
		query.Count(&total)

		var list []models.StockLog
		query.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&list)
		c.JSON(http.StatusOK, models.PageResult{Code: 0, Msg: "ok", Count: total, Data: list})
	}
}

// AdminVerifyPickupCode POST /admin/orders/verify
func AdminVerifyPickupCode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PickupCode string `json:"pickup_code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.PickupCode) == "" {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "请输入取货码"})
			return
		}

		var order models.Order
		if db.Where("pickup_code = ? AND status = 2", strings.TrimSpace(req.PickupCode)).First(&order).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "没有找到待取货订单"})
			return
		}

		now := time.Now()
		order.Status = 3
		order.VerifyBy = c.GetString("username")
		order.VerifyTime = &now
		if order.VerifyBy == "" {
			order.VerifyBy = "admin"
		}
		db.Save(&order)
		db.Create(&models.Message{
			UserID:  order.UserID,
			Title:   "订单已核销",
			Content: "订单 " + order.OrderNo + " 已完成取货核销",
			Type:    1,
		})

		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "核销成功", Data: order})
	}
}

func applyCouponDiscount(total float64, coupon models.UserCoupon) float64 {
	if total < coupon.Threshold {
		return 0
	}
	return math.Min(coupon.Amount, total)
}
