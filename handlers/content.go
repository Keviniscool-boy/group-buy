package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"zhixiang-group-buying/models"
)

// ─── 自提门店 ──────────────────────────────────

// ListStores GET /admin/stores /wx/stores
func ListStores(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		var total int64
		db.Model(&models.Store{}).Count(&total)

		var list []models.Store
		db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&list)

		c.JSON(http.StatusOK, models.PageResult{Code: 0, Msg: "ok", Count: total, Data: list})
	}
}

// CreateStore POST /admin/stores
func CreateStore(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var store models.Store
		if err := c.ShouldBindJSON(&store); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}
		db.Create(&store)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "新增成功"})
	}
}

// UpdateStore PUT /admin/stores/:id
func UpdateStore(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var store models.Store
		if db.First(&store, id).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "门店不存在"})
			return
		}
		c.ShouldBindJSON(&store)
		db.Save(&store)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "修改成功"})
	}
}

// DeleteStore DELETE /admin/stores/:id
func DeleteStore(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.Store{}, id)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "删除成功"})
	}
}

// WxListStores GET /wx/stores
func WxListStores(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.Store
		db.Where("status = 1").Order("id asc").Find(&list)
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: list})
	}
}

// ─── 站内消息 ──────────────────────────────────

// ListMessages GET /wx/messages
func ListMessages(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var list []models.Message
		db.Where("user_id = ? OR user_id = 0", userID).Order("id desc").Find(&list)
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: list})
	}
}

// ReadMessage PUT /wx/messages/:id/read
func ReadMessage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		db.Model(&models.Message{}).Where("id = ? AND (user_id = ? OR user_id = 0)", id, userID).Update("is_read", 1)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "已标记已读"})
	}
}

// CreateMessage POST /admin/messages
func CreateMessage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var msg models.Message
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}
		db.Create(&msg)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "发送成功"})
	}
}

// ─── 广告轮播图 ──────────────────────────────────

// ListAds GET /admin/ads /wx/ads
func ListAds(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.Ads
		db.Where("status = 1").Order("sort asc").Find(&list)
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: list})
	}
}

// AdminListAds GET /admin/ads/all
func AdminListAds(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		var total int64
		db.Model(&models.Ads{}).Count(&total)

		var list []models.Ads
		db.Order("sort asc").Offset((page - 1) * limit).Limit(limit).Find(&list)

		c.JSON(http.StatusOK, models.PageResult{Code: 0, Msg: "ok", Count: total, Data: list})
	}
}

// CreateAds POST /admin/ads
func CreateAds(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ads models.Ads
		if err := c.ShouldBindJSON(&ads); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}
		db.Create(&ads)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "新增成功"})
	}
}

// UpdateAds PUT /admin/ads/:id
func UpdateAds(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var ads models.Ads
		if db.First(&ads, id).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "广告不存在"})
			return
		}
		c.ShouldBindJSON(&ads)
		db.Save(&ads)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "修改成功"})
	}
}

// DeleteAds DELETE /admin/ads/:id
func DeleteAds(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.Ads{}, id)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "删除成功"})
	}
}

// ─── 公告 ──────────────────────────────────

// ListAnnouncements GET /admin/announcements /wx/announcements
func ListAnnouncements(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.Announcement
		db.Where("status = 1").Order("id desc").Find(&list)
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: list})
	}
}

// AdminListAnnouncements GET /admin/announcements/all
func AdminListAnnouncements(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		var total int64
		db.Model(&models.Announcement{}).Count(&total)

		var list []models.Announcement
		db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&list)

		c.JSON(http.StatusOK, models.PageResult{Code: 0, Msg: "ok", Count: total, Data: list})
	}
}

// CreateAnnouncement POST /admin/announcements
func CreateAnnouncement(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var a models.Announcement
		if err := c.ShouldBindJSON(&a); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}
		db.Create(&a)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "发布成功"})
	}
}

// UpdateAnnouncement PUT /admin/announcements/:id
func UpdateAnnouncement(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var a models.Announcement
		if db.First(&a, id).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "公告不存在"})
			return
		}
		c.ShouldBindJSON(&a)
		db.Save(&a)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "修改成功"})
	}
}

// DeleteAnnouncement DELETE /admin/announcements/:id
func DeleteAnnouncement(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.Announcement{}, id)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "删除成功"})
	}
}
