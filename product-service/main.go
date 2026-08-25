package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	nacosConfig "rxtcloud/common/config"
	svcConfig "rxtcloud/product-service/config"
	"rxtcloud/product-service/router"
	"rxtcloud/product-service/service"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("========================================")
	log.Println("  商品服务 (Product-Service) 启动中...")
	log.Println("========================================")

	cfg := svcConfig.LoadConfig()
	log.Printf("[Config] 服务名: %s, 端口: %s", cfg.ServiceName, cfg.ServicePort)

	svcConfig.InitDB(cfg.MySQLDSN)

	// 初始化敏感词缓存
	if err := service.InitSensitiveWords(); err != nil {
		log.Printf("[Warning] 敏感词初始化失败: %v", err)
	}

	nacosCfg := nacosConfig.DefaultNacosConfig()
	namingClient, err := nacosConfig.InitNamingClient(nacosCfg)
	if err != nil {
		log.Printf("[Nacos] 初始化失败（服务仍可运行）: %v", err)
	} else {
		port := nacosConfig.ParsePort(cfg.ServicePort)
		if err := nacosConfig.RegisterService(namingClient, cfg.ServiceName, cfg.ServiceIP, port); err != nil {
			log.Printf("[Nacos] 注册失败: %v", err)
		}
		defer nacosConfig.DeregisterService(namingClient, cfg.ServiceName, cfg.ServiceIP, port)
	}

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	router.InitRouter(r)

	srv := &http.Server{Addr: ":" + cfg.ServicePort, Handler: r}

	go func() {
		log.Printf("[HTTP] 商品服务启动于 :%s", cfg.ServicePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[HTTP] 启动失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[Shutdown] 正在关闭商品服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[Shutdown] 强制关闭: %v", err)
	}
	log.Println("[Shutdown] 商品服务已关闭")
}
