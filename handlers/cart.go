package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"zhixiang-group-buying/config"
	"zhixiang-group-buying/models"
)

// WxAddToCart POST /wx/cart
func WxAddToCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		var req struct {
			CommodityID uint `json:"commodity_id"`
			Quantity    int  `json:"quantity"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}

		// 获取商品资讯
		var com models.Commodity
		if db.First(&com, req.CommodityID).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "商品不存在"})
			return
		}
		if com.Status != 1 {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "商品已下架"})
			return
		}
		if com.SaleStartTime != nil && com.SaleStartTime.After(time.Now()) {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "商品尚未开售，可先订阅提醒"})
			return
		}
		if req.Quantity <= 0 {
			req.Quantity = 1
		}
		if com.Stock < req.Quantity {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "商品库存不足"})
			return
		}

		// 找到或建立购物车
		var cart models.ShoppingCart
		db.Where("user_id = ?", userID).First(&cart)
		if cart.ID == 0 {
			cart = models.ShoppingCart{UserID: userID}
			db.Create(&cart)
		}

		// 检查购物车中是否已有此商品
		var existItem models.CartItem
		if db.Where("cart_id = ? AND commodity_id = ?", cart.ID, req.CommodityID).First(&existItem).Error == nil {
			if com.Stock < existItem.Quantity+req.Quantity {
				c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "商品库存不足"})
				return
			}
			existItem.Quantity += req.Quantity
			if existItem.Quantity <= 0 {
				db.Delete(&existItem)
			} else {
				db.Save(&existItem)
			}
		} else {
			price := com.Price
			db.Create(&models.CartItem{
				CartID:        cart.ID,
				CommodityID:   com.ID,
				CommodityName: com.Name,
				Price:         price,
				Quantity:      req.Quantity,
				Image:         com.Image,
				Checked:       1,
			})
		}

		config.CacheDel(c, fmt.Sprintf("zhixiang:cart:%d", userID))
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "已加入购物车"})
	}
}

// WxListCart GET /wx/cart
func WxListCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		cacheKey := fmt.Sprintf("zhixiang:cart:%d", userID)
		if cached, ok := config.CacheGet(c, cacheKey); ok {
			var items []models.CartItem
			if json.Unmarshal([]byte(cached), &items) == nil {
				c.JSON(http.StatusOK, models.Result{Code: 0, Data: items})
				return
			}
		}

		var cart models.ShoppingCart
		db.Where("user_id = ?", userID).First(&cart)
		if cart.ID == 0 {
			c.JSON(http.StatusOK, models.Result{Code: 0, Data: []models.CartItem{}})
			return
		}

		var items []models.CartItem
		db.Where("cart_id = ?", cart.ID).Order("id desc").Find(&items)

		if data, err := json.Marshal(items); err == nil {
			config.CacheSet(c, cacheKey, data, 30*time.Minute)
		}

		c.JSON(http.StatusOK, models.Result{Code: 0, Data: items})
	}
}

// WxUpdateCartItem PUT /wx/cart/:id
func WxUpdateCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")

		var cart models.ShoppingCart
		db.Where("user_id = ?", userID).First(&cart)

		var item models.CartItem
		if db.Where("id = ? AND cart_id = ?", id, cart.ID).First(&item).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "购物车项不存在"})
			return
		}

		var req struct {
			Quantity *int `json:"quantity"`
			Checked  *int `json:"checked"`
		}
		c.ShouldBindJSON(&req)

		if req.Quantity != nil {
			if *req.Quantity <= 0 {
				db.Delete(&item)
				c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "已删除"})
				return
			}
			item.Quantity = *req.Quantity
		}
		if req.Checked != nil {
			item.Checked = *req.Checked
		}
		db.Save(&item)
		config.CacheDel(c, fmt.Sprintf("zhixiang:cart:%d", userID))
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "修改成功"})
	}
}

// WxDeleteCartItem DELETE /wx/cart/:id
func WxDeleteCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")

		var cart models.ShoppingCart
		db.Where("user_id = ?", userID).First(&cart)

		db.Where("id = ? AND cart_id = ?", id, cart.ID).Delete(&models.CartItem{})
		config.CacheDel(c, fmt.Sprintf("zhixiang:cart:%d", userID))
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "删除成功"})
	}
}

// WxCheckAllCart PUT /wx/cart/check-all
func WxCheckAllCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		var req struct {
			Checked int `json:"checked"`
		}
		c.ShouldBindJSON(&req)

		var cart models.ShoppingCart
		db.Where("user_id = ?", userID).First(&cart)
		db.Model(&models.CartItem{}).Where("cart_id = ?", cart.ID).Update("checked", req.Checked)
		config.CacheDel(c, fmt.Sprintf("zhixiang:cart:%d", userID))

		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "操作成功"})
	}
}

// WxClearCart DELETE /wx/cart/clear
func WxClearCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var cart models.ShoppingCart
		db.Where("user_id = ?", userID).First(&cart)
		db.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{})
		config.CacheDel(c, fmt.Sprintf("zhixiang:cart:%d", userID))
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "购物车已清空"})
	}
}
