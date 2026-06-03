package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"zhixiang-group-buying/models"
)

// ─── 商品分类 ──────────────────────────────────

// ListCategories GET /admin/categories /wx/categories
func ListCategories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.CommodityCategory
		db.Order("sort asc").Find(&list)
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: list})
	}
}

// CreateCategory POST /admin/categories
func CreateCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cat models.CommodityCategory
		if err := c.ShouldBindJSON(&cat); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}
		db.Create(&cat)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "新增成功"})
	}
}

// UpdateCategory PUT /admin/categories/:id
func UpdateCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var cat models.CommodityCategory
		if db.First(&cat, id).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "分类不存在"})
			return
		}
		c.ShouldBindJSON(&cat)
		db.Save(&cat)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "修改成功"})
	}
}

// DeleteCategory DELETE /admin/categories/:id
func DeleteCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.CommodityCategory{}, id)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "删除成功"})
	}
}

// ─── 商品管理（后台）────────────────────────────

// AdminListCommodities GET /admin/commodities
func AdminListCommodities(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		keyword := c.Query("keyword")
		categoryID := c.Query("category_id")

		query := db.Model(&models.Commodity{})
		if keyword != "" {
			query = query.Where("name LIKE ?", "%"+keyword+"%")
		}
		if categoryID != "" {
			query = query.Where("category_id = ?", categoryID)
		}

		var total int64
		query.Count(&total)

		var list []models.Commodity
		query.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&list)

		c.JSON(http.StatusOK, models.PageResult{
			Code: 0, Msg: "ok", Count: total, Data: list,
		})
	}
}

// CreateCommodity POST /admin/commodities
func CreateCommodity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var com models.Commodity
		if err := c.ShouldBindJSON(&com); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}
		// 填入分类名称
		var cat models.CommodityCategory
		if db.First(&cat, com.CategoryID).Error == nil {
			com.CategoryName = cat.Name
		}
		db.Create(&com)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "发布成功"})
	}
}

// UpdateCommodity PUT /admin/commodities/:id
func UpdateCommodity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var com models.Commodity
		if db.First(&com, id).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "商品不存在"})
			return
		}
		oldStock := com.Stock
		c.ShouldBindJSON(&com)
		// 更新分类名称
		var cat models.CommodityCategory
		if db.First(&cat, com.CategoryID).Error == nil {
			com.CategoryName = cat.Name
		}
		db.Save(&com)
		if com.Stock != oldStock {
			typ := "admin"
			remark := "后台调整库存"
			if com.Stock > oldStock {
				typ = "restock"
				remark = "后台补货入库"
			}
			CreateStockLog(db, com, com.Stock-oldStock, oldStock, com.Stock, typ, com.ID, remark, c.GetString("username"))
		}
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "修改成功"})
	}
}

// ToggleCommodity PUT /admin/commodities/:id/toggle
func ToggleCommodity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var com models.Commodity
		if db.First(&com, id).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "商品不存在"})
			return
		}
		com.Status = 1 - com.Status
		db.Save(&com)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "操作成功"})
	}
}

// DeleteCommodity DELETE /admin/commodities/:id
func DeleteCommodity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.Commodity{}, id)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "删除成功"})
	}
}

// ─── 商品浏览（微信端）─────────────────────────

// WxListCommodities GET /wx/commodities
func WxListCommodities(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryID := c.Query("category_id")
		keyword := c.Query("keyword")
		isGroupon := c.Query("is_groupon") // 1=团购商品

		query := db.Where("status = 1")
		if categoryID != "" {
			query = query.Where("category_id = ?", categoryID)
		}
		if keyword != "" {
			query = query.Where("name LIKE ?", "%"+keyword+"%")
		}
		if isGroupon == "1" {
			query = query.Where("is_groupon = 1")
		}

		var list []models.Commodity
		query.Order("id desc").Find(&list)

		c.JSON(http.StatusOK, models.Result{Code: 0, Data: list})
	}
}

// WxGetCommodity GET /wx/commodities/:id
func WxGetCommodity(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var com models.Commodity
		if db.First(&com, id).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "商品不存在"})
			return
		}
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: com})
	}
}
