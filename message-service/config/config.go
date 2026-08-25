package config

import "os"

// ServiceConfig 消息服务专属配置
type ServiceConfig struct {
	ServiceName string
	ServicePort string
	ServiceIP   string
	MySQLDSN    string
}

func LoadConfig() ServiceConfig {
	return ServiceConfig{
		ServiceName: getEnv("SERVICE_NAME", "message-service"),
		ServicePort: getEnv("SERVICE_PORT", "8082"),
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
