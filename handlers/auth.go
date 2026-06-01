package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"zhixiang-group-buying/middleware"
	"zhixiang-group-buying/models"
)

// 简易内存验证码快取
var captchaStore sync.Map

// ─── 后台登入 ──────────────────────────────────

// AdminLogin POST /admin/login
func AdminLogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}

		// 验证码校验（演示环境直接回传验证码，但登录必须填写）
		v, ok := captchaStore.Load(req.Username + "_captcha")
		if !ok {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "请先获取验证码"})
			return
		}
		if v.(string) != req.Captcha {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "验证码错误"})
			return
		}
		captchaStore.Delete(req.Username + "_captcha")

		var admin models.Admin
		if err := db.Where("username = ?", req.Username).First(&admin).Error; err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "帐号或密码错误"})
			return
		}

		if admin.Status == 0 {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "帐号已被停用"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "帐号或密码错误"})
			return
		}

		token, _ := middleware.GenerateToken(admin.ID, admin.Username, admin.RoleID, admin.RoleName, true)

		c.JSON(http.StatusOK, models.Result{
			Code: 0,
			Msg:  "登入成功",
			Data: models.LoginResult{
				Token: token,
				ID:    admin.ID,
				Name:  admin.Username,
				Role:  admin.RoleName,
			},
		})
	}
}

// AdminRegister POST /admin/register
func AdminRegister(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var adminCount int64
		db.Model(&models.Admin{}).Count(&adminCount)
		if adminCount > 0 {
			c.JSON(http.StatusOK, models.Result{Code: 403, Msg: "请登录后台后新增管理员"})
			return
		}
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}

		// 检查用户名唯一性
		var exist models.Admin
		if db.Where("username = ?", req.Username).First(&exist).Error == nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "用户名已存在"})
			return
		}

		hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

		admin := models.Admin{
			Username: req.Username,
			Password: string(hashed),
			RoleName: "业务管理员",
			Status:   1,
		}
		db.Create(&admin)

		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "注册成功"})
	}
}

// GetCaptcha GET /admin/captcha
func GetCaptcha(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Query("username")
		if username == "" {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "请提供用户名"})
			return
		}
		code := fmt.Sprintf("%04d", rand.Intn(10000))
		captchaStore.Store(username+"_captcha", code)
		// 演示环境直接回传验证码
		c.JSON(http.StatusOK, models.Result{Code: 0, Msg: "验证码已生成", Data: gin.H{"captcha": code}})
	}
}

// ─── 微信端登入 ──────────────────────────────────

// WxLogin POST /wx/login
func WxLogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.WxLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, models.Result{Code: 400, Msg: "参数错误"})
			return
		}

		// 模拟：以 code 作为 openid
		openID := req.Code
		if openID == "" {
			openID = fmt.Sprintf("wx_openid_%d_%d", time.Now().UnixMilli(), rand.Intn(10000))
		}

		var user models.User
		err := db.Where("open_id = ?", openID).First(&user).Error

		if err != nil {
			// 新用户自动注册
			nickname := req.Nickname
			if nickname == "" {
				nickname = fmt.Sprintf("用户%04d", rand.Intn(10000))
			}
			user = models.User{
				OpenID:   openID,
				Nickname: nickname,
				Avatar:   req.Avatar,
				Status:   1,
			}
			db.Create(&user)
		}

		token, _ := middleware.GenerateToken(user.ID, user.Nickname, 0, "", false)

		c.JSON(http.StatusOK, models.Result{
			Code: 0,
			Msg:  "登入成功",
			Data: models.LoginResult{
				Token: token,
				ID:    user.ID,
				Name:  user.Nickname,
			},
		})
	}
}
