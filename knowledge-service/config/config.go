package config

import "os"

// ServiceConfig 知识服务专属配置
type ServiceConfig struct {
	ServiceName string // Nacos 服务名
	ServicePort string // 服务端口
	ServiceIP   string // 服务IP（空则自动获取）
	MySQLDSN    string // MySQL 连接字符串
}

// LoadConfig 加载配置（优先读取环境变量，否则使用默认值）
func LoadConfig() ServiceConfig {
	return ServiceConfig{
		ServiceName: getEnv("SERVICE_NAME", "knowledge-service"),
		ServicePort: getEnv("SERVICE_PORT", "8081"),
		ServiceIP:   getEnv("SERVICE_IP", ""),
		MySQLDSN:    getEnv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/gorxt_db?charset=utf8mb4&parseTime=True&loc=Local"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
