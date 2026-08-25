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
	svcConfig "rxtcloud/knowledge-service/config"
	"rxtcloud/knowledge-service/router"
	"rxtcloud/knowledge-service/service"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("========================================")
	log.Println("  知识服务 (Knowledge-Service) 启动中...")
	log.Println("========================================")

	// 1. 加载服务配置
	cfg := svcConfig.LoadConfig()
	log.Printf("[Config] 服务名: %s, 端口: %s", cfg.ServiceName, cfg.ServicePort)

	// 2. 初始化数据库
	svcConfig.InitDB(cfg.MySQLDSN)

	// 初始化敏感词缓存
	if err := service.InitSensitiveWords(); err != nil {
		log.Printf("[Warning] 敏感词初始化失败: %v", err)
	}

	// 3. 初始化 Nacos 注册中心
	nacosCfg := nacosConfig.DefaultNacosConfig()
	namingClient, err := nacosConfig.InitNamingClient(nacosCfg)
	if err != nil {
		log.Printf("[Nacos] 初始化失败（服务仍可运行）: %v", err)
	} else {
		// 注册服务到 Nacos
		port := nacosConfig.ParsePort(cfg.ServicePort)
		if err := nacosConfig.RegisterService(namingClient, cfg.ServiceName, cfg.ServiceIP, port); err != nil {
			log.Printf("[Nacos] 注册失败: %v", err)
		}

		// 优雅退出时注销服务
		defer nacosConfig.DeregisterService(namingClient, cfg.ServiceName, cfg.ServiceIP, port)
	}

	// 4. 创建 Gin 引擎
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 禁用尾斜杠/固定路径自动重定向，避免 /api/knowledge 路由与
	// r.Group("/knowledge") 分组发生 RedirectTrailingSlash 死循环
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	// 5. 注册路由
	router.InitRouter(r)

	// 6. 启动 HTTP 服务
	srv := &http.Server{
		Addr:    ":" + cfg.ServicePort,
		Handler: r,
	}

	go func() {
		log.Printf("[HTTP] 知识服务启动于 :%s", cfg.ServicePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[HTTP] 启动失败: %v", err)
		}
	}()

	// 7. 等待中断信号，优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[Shutdown] 正在关闭知识服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[Shutdown] 强制关闭: %v", err)
	}

	log.Println("[Shutdown] 知识服务已关闭")
}
