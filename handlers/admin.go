package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"zhixiang-group-buying/models"
)

// ─── 后台管理员 CRUD ─────────────────────────────

// ListAdmins GET /admin/admins
func ListAdmins(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		var total int64
		db.Model(&models.Admin{}).Count(&total)

		var list []models.Admin
		db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&list)

		c.JSON(http.StatusOK, models.PageResult{Code: 0, Msg: "ok", Count: total, Data: list})
	}
}

// CreateAdmin POST /admin/admins
func CreateAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
			RoleID   uint   `json:"role_id"`
			RoleName string `json:"role_name"`
			Status   int    `json:"status"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}
		var exist models.Admin
		if db.Where("username = ?", req.Username).First(&exist).Error == nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "管理员账号已存在"})
			return
		}
		if req.RoleName == "" && req.RoleID > 0 {
			var role models.Role
			if db.First(&role, req.RoleID).Error == nil {
				req.RoleName = role.Name
			}
		}
		if req.RoleName == "" {
			req.RoleName = "业务管理员"
		}
		if req.Status != 0 {
			req.Status = 1
		}
		hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		admin := models.Admin{
			Username: req.Username,
			Password: string(hashed),
			RoleID:   req.RoleID,
			RoleName: req.RoleName,
			Status:   req.Status,
		}
		db.Create(&admin)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "新增成功", Data: admin})
	}
}

// UpdateAdmin PUT /admin/admins/:id
func UpdateAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var admin models.Admin
		if db.First(&admin, id).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "管理员不存在"})
			return
		}
		var req struct {
			RoleID   uint   `json:"role_id"`
			RoleName string `json:"role_name"`
			Status   int    `json:"status"`
			Password string `json:"password"`
		}
		c.ShouldBindJSON(&req)

		if req.RoleName != "" {
			admin.RoleID = req.RoleID
			admin.RoleName = req.RoleName
		}
		admin.Status = req.Status
		if req.Password != "" {
			hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			admin.Password = string(hashed)
		}
		db.Save(&admin)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "修改成功"})
	}
}

// DeleteAdmin DELETE /admin/admins/:id
func DeleteAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if strconv.FormatUint(uint64(c.GetUint("userID")), 10) == id {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "不能删除当前登录管理员"})
			return
		}
		db.Delete(&models.Admin{}, id)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "删除成功"})
	}
}

// ─── 前台用户管理 ──────────────────────────────────

// ListUsers GET /admin/users
func ListUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		keyword := c.Query("keyword")

		query := db.Model(&models.User{})
		if keyword != "" {
			query = query.Where("nickname LIKE ? OR phone LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}

		var total int64
		query.Count(&total)

		var list []models.User
		query.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&list)

		c.JSON(http.StatusOK, models.PageResult{Code: 0, Msg: "ok", Count: total, Data: list})
	}
}

// DeleteUser DELETE /admin/users/:id
func DeleteUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var carts []models.ShoppingCart
		db.Where("user_id = ?", id).Find(&carts)
		for _, cart := range carts {
			db.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{})
		}
		db.Where("user_id = ?", id).Delete(&models.ShoppingCart{})
		db.Delete(&models.User{}, id)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "删除成功"})
	}
}

// ToggleUser PUT /admin/users/:id/toggle
func ToggleUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var user models.User
		if db.First(&user, id).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "用户不存在"})
			return
		}
		user.Status = 1 - user.Status
		db.Save(&user)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "操作成功"})
	}
}

// ─── 微信端个人中心 ──────────────────────────────────

// WxUpdateProfile PUT /wx/profile
func WxUpdateProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var user models.User
		if db.First(&user, userID).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "用户不存在"})
			return
		}
		var req struct {
			Nickname string `json:"nickname"`
			Phone    string `json:"phone"`
			Avatar   string `json:"avatar"`
		}
		c.ShouldBindJSON(&req)
		if req.Nickname != "" {
			user.Nickname = req.Nickname
		}
		if req.Phone != "" {
			user.Phone = req.Phone
		}
		if req.Avatar != "" {
			user.Avatar = req.Avatar
		}
		db.Save(&user)
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "修改成功", Data: user})
	}
}

// WxGetProfile GET /wx/profile
func WxGetProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var user models.User
		if db.First(&user, userID).Error != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "用户不存在"})
			return
		}
		c.JSON(http.StatusOK, models.Result{Code: 0, Data: user})
	}
}

// ─── Dashboard 统计 ──────────────────────────────────

// AdminDashboard GET /admin/dashboard
func AdminDashboard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userCount, orderCount, commodityCount, todayOrderCount int64
		db.Model(&models.User{}).Count(&userCount)
		db.Model(&models.Order{}).Count(&orderCount)
		db.Model(&models.Commodity{}).Count(&commodityCount)
		db.Model(&models.Order{}).Where("DATE(created_at) = CURDATE()").Count(&todayOrderCount)

		// 今日销售额
		var todaySales float64
		db.Model(&models.Order{}).Where("DATE(created_at) = CURDATE() AND status IN (1,2,3)").Select("COALESCE(SUM(total_amount), 0)").Scan(&todaySales)

		c.JSON(http.StatusOK, models.Result{Code: 0, Data: gin.H{
			"user_count":        userCount,
			"order_count":       orderCount,
			"commodity_count":   commodityCount,
			"today_order_count": todayOrderCount,
			"today_sales":       todaySales,
		}})
	}
}
