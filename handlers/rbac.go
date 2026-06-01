package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"zhixiang-group-buying/models"
)

// ─── 角色管理 ──────────────────────────────────

// ListRoles GET /admin/roles
func ListRoles(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.Role
		db.Order("id asc").Find(&list)
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: list})
	}
}

// CreateRole POST /admin/roles
func CreateRole(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var role models.Role
		if err := c.ShouldBindJSON(&role); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}
		db.Create(&role)
		// 建立权限关联
		if role.Menus != "" {
			// menus 格式: "1,2,3"
			// 此处可扩展写入 cs_authority 表
		}
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "新增成功"})
	}
}

// UpdateRole PUT /admin/roles/:id
func UpdateRole(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var role models.Role
		if db.First(&role, id).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "角色不存在"})
			return
		}
		c.ShouldBindJSON(&role)
		db.Save(&role)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "修改成功"})
	}
}

// DeleteRole DELETE /admin/roles/:id
func DeleteRole(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.Role{}, id)
		db.Where("role_id = ?", id).Delete(&models.Authority{})
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "删除成功"})
	}
}

// ─── 菜单管理 ──────────────────────────────────

// ListMenus GET /admin/menus
func ListMenus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.Menu
		db.Order("sort asc").Find(&list)
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: list})
	}
}

// CreateMenu POST /admin/menus
func CreateMenu(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var menu models.Menu
		if err := c.ShouldBindJSON(&menu); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}
		db.Create(&menu)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "新增成功"})
	}
}

// DeleteMenu DELETE /admin/menus/:id
func DeleteMenu(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.Menu{}, id)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "删除成功"})
	}
}

// ─── 权限管理 ──────────────────────────────────

// ListAuthorities GET /admin/authorities
func ListAuthorities(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID := c.Query("role_id")
		var list []models.Authority
		query := db.Model(&models.Authority{})
		if roleID != "" {
			query = query.Where("role_id = ?", roleID)
		}
		query.Find(&list)
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: list})
	}
}

// SaveAuthorities POST /admin/authorities
func SaveAuthorities(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RoleID  uint   `json:"role_id"`
			MenuIDs []uint `json:"menu_ids"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}

		// 删除旧权限
		db.Where("role_id = ?", req.RoleID).Delete(&models.Authority{})

		// 新增权限
		for _, menuID := range req.MenuIDs {
			db.Create(&models.Authority{RoleID: req.RoleID, MenuID: menuID})
		}

		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "权限保存成功"})
	}
}

// ─── 团长销量查询 ──────────────────────────────────

// LeaderSales GET /admin/leader-sales
func LeaderSales(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 回传各门店的订单统计（模拟团长销量）
		type StoreSales struct {
			StoreID     uint    `json:"store_id"`
			StoreName   string  `json:"store_name"`
			OrderCount  int64   `json:"order_count"`
			TotalSales  float64 `json:"total_sales"`
		}
		var results []StoreSales
		db.Model(&models.Order{}).
			Select("store_id, store_name, COUNT(*) as order_count, COALESCE(SUM(total_amount),0) as total_sales").
			Where("store_id > 0").
			Group("store_id, store_name").
			Scan(&results)

		c.JSON(http.StatusOK, models.Result{Code: 0, Data: results})
	}
}

// ─── 团购开团提醒 ──────────────────────────────────

// WxSubscribeGroupon POST /wx/groupon/subscribe
func WxSubscribeGroupon(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		var req struct {
			CommodityID uint `json:"commodity_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}

		// 检查是否已订阅
		var exist models.GrouponRemind
		if db.Where("user_id = ? AND commodity_id = ?", userID, req.CommodityID).First(&exist).Error == nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "已订阅过该商品"})
			return
		}

		var com models.Commodity
		db.First(&com, req.CommodityID)

		db.Create(&models.GrouponRemind{
			UserID:        userID,
			CommodityID:   req.CommodityID,
			CommodityName: com.Name,
		})

		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "订阅成功，开团时将收到提醒"})
	}
}

// WxListGrouponReminds GET /wx/groupon/reminds
func WxListGrouponReminds(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var list []models.GrouponRemind
		db.Where("user_id = ?", userID).Order("id desc").Find(&list)
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: list})
	}
}

// WxDeleteGrouponRemind DELETE /wx/groupon/reminds/:id
func WxDeleteGrouponRemind(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.GrouponRemind{})
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "已取消订阅"})
	}
}
