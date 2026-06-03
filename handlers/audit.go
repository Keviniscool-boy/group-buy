package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"zhixiang-group-buying/models"
)

// AdminAuditLog records mutating admin operations for audit and demo review.
func AdminAuditLog(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete {
			return
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		if strings.Contains(path, "/operation-logs") {
			return
		}

		log := models.OperationLog{
			AdminID:   c.GetUint("userID"),
			Username:  c.GetString("username"),
			RoleName:  c.GetString("roleName"),
			Method:    method,
			Path:      path,
			Action:    actionName(method, path),
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Status:    c.Writer.Status(),
			Duration:  time.Since(start).Milliseconds(),
		}
		if len(log.UserAgent) > 255 {
			log.UserAgent = log.UserAgent[:255]
		}
		db.Create(&log)
	}
}

func actionName(method, path string) string {
	if method == http.MethodPost {
		return "新增/提交"
	}
	if method == http.MethodPut {
		return "修改/处理"
	}
	if method == http.MethodDelete {
		return "删除/取消"
	}
	return "访问"
}

// ListOperationLogs GET /admin/operation-logs
func ListOperationLogs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		keyword := c.Query("keyword")
		method := c.Query("method")

		query := db.Model(&models.OperationLog{})
		if keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where("username LIKE ? OR path LIKE ? OR action LIKE ?", like, like, like)
		}
		if method != "" {
			query = query.Where("method = ?", method)
		}

		var total int64
		query.Count(&total)

		var list []models.OperationLog
		query.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&list)
		c.JSON(http.StatusOK, models.PageResult{Code: 0, Msg: "ok", Count: total, Data: list})
	}
}
