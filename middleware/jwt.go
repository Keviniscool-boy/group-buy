package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"zhixiang-group-buying/config"
)

// Claims 自订 JWT 内容
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RoleID   uint   `json:"role_id"`
	RoleName string `json:"role_name"`
	IsAdmin  bool   `json:"is_admin"` // true=后台管理者 false=前台用户
	jwt.RegisteredClaims
}

// AdminAuth 后台管理者 JWT 验证中介
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			// 也尝试从 query / cookie 取
			tokenStr = c.Query("token")
		}
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "请先登入"})
			c.Abort()
			return
		}

		claims, err := parseToken(tokenStr)
		if err != nil || !claims.IsAdmin {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token无效或已过期"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roleID", claims.RoleID)
		c.Set("roleName", claims.RoleName)
		c.Next()
	}
}

// WxAuth 微信小程式用户 JWT 验证中介
func WxAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		if tokenStr == "" {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "请先登入"})
			c.Abort()
			return
		}

		claims, err := parseToken(tokenStr)
		if err != nil || claims.IsAdmin {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token无效或已过期"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// GenerateToken 签发 Token
func GenerateToken(userID uint, username string, roleID uint, roleName string, isAdmin bool) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RoleName: roleName,
		IsAdmin:  isAdmin,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

// parseToken 解析 Token
func parseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
