package middleware

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"rxtcloud/common/client"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 角色：API 网关反向代理 — 对标 Spring Cloud Gateway / Zuul
//
// 本文件是 API 网关的核心实现，负责将前端请求透明转发到目标微服务。
//
// 工作流程：
//
//   浏览器请求
//   GET /api/products
//        │
//        ▼
//   ┌─────────────────┐    ① 路由器匹配路由规则
//   │ common-service  │       api.Any("/products", ProxyTo("product-service"))
//   │   (网关:8080)    │
//   └────────┬────────┘    ② ProxyTo 中间件被调用
//            │
//            ▼              ③ discovery.GetServiceURL("product-service")
//   ┌────────────────────┐     → Nacos 返回 127.0.0.1:8084
//   │  httputil.ReverseProxy  │  → 或本地回退 127.0.0.1:8084
//   └────────┬───────────┘
//            │              ④ 创建反向代理，将请求原样转发
//            ▼
//   ┌─────────────────┐    ⑤ product-service 处理请求，返回响应
//   │ product-service │    ⑥ 响应原路返回给浏览器
//   │    (:8084)       │
//   └─────────────────┘
//
// 对前端完全透明：前端只需要知道 common-service(:8080)，
// 不知道后面有几个微服务、各自端口是多少。
// 这是微服务架构中 Gateway 的核心价值。
// ============================================================

// ProxyTo 创建反向代理中间件，将请求透明转发到指定微服务
//
// 参数：
//   serviceName: Nacos 服务名（如 "knowledge-service", "product-service"）
//
// 返回：
//   gin.HandlerFunc: 可以直接注册到 Gin 路由的中间件函数
//
// 使用方式（在 router/router.go 中）：
//   proxyKnowledge := middleware.ProxyTo("knowledge-service")
//   api.Any("/knowledge/*path", proxyKnowledge)
//
// 特性：
//   - 透明转发：请求的 Method、Header、Body、Query 全部原样传递
//   - 服务发现：通过 Nacos 动态获取目标地址
//   - 容错回退：Nacos 不可用时走本地 fallback
//   - 错误处理：超时/连接失败返回 502，服务不存在返回 503
func ProxyTo(serviceName string) gin.HandlerFunc {
	// 获取全局唯一的服务发现单例
	discovery := client.GetDiscovery()

	return func(c *gin.Context) {
		// ① 通过 Nacos 服务发现获取目标服务地址
		//    结果示例: "http://127.0.0.1:8084"
		targetURL, err := discovery.GetServiceURL(serviceName)
		if err != nil {
			// 503 Service Unavailable：目标服务不可用
			log.Printf("[Proxy] 服务 %s 不可用: %v", serviceName, err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
			c.Abort()
			return
		}

		// ② 将地址字符串解析为 url.URL
		target, err := url.Parse(targetURL)
		if err != nil {
			// 500 Internal Server Error：配置错误
			c.JSON(http.StatusInternalServerError, gin.H{"error": "proxy error"})
			c.Abort()
			return
		}

		// ③ 创建 Go 标准库的单目标反向代理
		//    httputil.ReverseProxy 自动处理：
		//    - 拷贝请求头/响应头/状态码
		//    - 重写 Host 头为目标地址
		//    - 处理 WebSocket 升级（如有需要）
		proxy := httputil.NewSingleHostReverseProxy(target)

		// ④ 自定义错误处理：下游连接失败返回 502 Bad Gateway
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[Proxy] %s 转发失败: %v", serviceName, err)
			w.WriteHeader(http.StatusBadGateway)
		}

		// ⑤ 执行代理转发
		//    c.Request 会被修改（Host/Path 等），但原始请求不会被复制
		//    响应直接写入 c.Writer，返回给前端
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
