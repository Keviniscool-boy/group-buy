package config

var (
	// DB 资料库连线参数
	DBUser     = "root"
	DBPassword = "123456"
	DBHost     = "127.0.0.1"
	DBPort     = "3306"
	DBName     = "zhixiang_group_buying"
	DBCharset  = "utf8mb4"

	// JWT 密钥
	JWTSecret = "zhixiang-jwt-secret-key-2024"

	// 服务埠
	ServerPort = ":8080"
)

// GetDSN 回传 MySQL 连线字串
func GetDSN() string {
	return DBUser + ":" + DBPassword + "@tcp(" + DBHost + ":" + DBPort + ")/" + DBName + "?charset=" + DBCharset + "&parseTime=true&loc=Local"
}
