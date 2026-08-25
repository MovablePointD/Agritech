package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rxtcloud/common/config"
	"rxtcloud/common/router"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 角色：common-service 启动入口 — API 网关 + 公共模块
//
// 本文件是 common-service 的 main 函数，承担两个角色：
//   ① API 网关：将前端请求路由到对应微服务（通过 ProxyTo）
//   ② 公共模块：自身也承载未拆分的模块（用户/地址/专家/农村/上传）
//
// 启动流程（共 4 步）：
//
//   ① config.InitDB()
//      └─ 连接 MySQL 数据库
//
//   ② → Nacos 服务注册
//      └─ 向 Nacos Server 注册 "common-service@127.0.0.1:8080"
//
//   ③ → router.InitRouter()
//      └─ 注册自身路由 + 代理路由
//
//   ④ → http.ListenAndServe(":8080")
//      └─ 启动 HTTP 服务，等待请求
//
//   ⑤ → 等待系统信号 (SIGINT/SIGTERM)
//      └─ Ctrl+C → 优雅关闭 → defer 注销 Nacos
// ============================================================

func main() {
	log.Println("========================================")
	log.Println("  公共服务 (Common-Service) 启动中...")
	log.Println("  模块: 用户/地址/专家/农村信息/农村事务/上传/敏感词/事务处理")
	log.Println("  角色: API网关 + 公共模块")
	log.Println("========================================")

	// ── 第一步：初始化数据库连接 ──
	// common-service 管理用户/地址/专家/农村等核心数据表
	config.InitDB()

	// ── 第二步：向 Nacos 注册中心注册本服务 ──
	//
	// 为什么需要注册？
	//   其他微服务（如 knowledge-service）需要调用 common-service 的接口时，
	//   通过 Nacos 发现 common-service 的地址，而不是硬编码 IP。
	//
	// 容错设计：
	//   如果 Nacos 初始化或注册失败，会在日志中打印错误，
	//   但不会 panic 退出——服务仍然可以通过本地回退（fallbackURL）工作。
	nacosCfg := config.DefaultNacosConfig()
	namingClient, err := config.InitNamingClient(nacosCfg)
	if err != nil {
		log.Printf("[Nacos] 初始化失败（服务仍可运行）: %v", err)
	} else {
		port := config.ParsePort("8080")
		// 注册到 Nacos：服务名 "common-service"，IP 自动获取 127.0.0.1
		if err := config.RegisterService(namingClient, "common-service", "", port); err != nil {
			log.Printf("[Nacos] 注册失败: %v", err)
		}
		// defer 确保程序退出时从 Nacos 注销，避免残留无效实例
		defer config.DeregisterService(namingClient, "common-service", "", port)
	}

	// ── 第三步：初始化路由 ──
	// router.InitRouter() 内部：
	//   先注册公共模块路由（用户/地址/专家等本地处理）
	//   再注册代理路由（knowledge/posts/products 等转发到微服务）
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := router.InitRouter()

	// ── 第四步：启动 HTTP 服务 ──
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Printf("[HTTP] 公共服务启动于 :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[HTTP] 启动失败: %v", err)
		}
	}()

	// ── 第五步：等待系统信号，优雅关闭 ──
	// 监听 Ctrl+C (SIGINT) 和 kill (SIGTERM) 信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 阻塞直到收到信号

	log.Println("[Shutdown] 正在关闭公共服务...")

	// 给 5 秒时间让正在处理的请求完成
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 平滑关闭：不接受新请求，等待已有请求处理完毕
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[Shutdown] 强制关闭: %v", err)
	}

	// 程序退出时 defer 的 DeregisterService 会执行，从 Nacos 注销
	log.Println("[Shutdown] 公共服务已关闭")
}
