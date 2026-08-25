package router

import (
	"rxtcloud/common/middleware"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 角色：API 网关路由表 — 对标 Spring Cloud Gateway 的路由配置
//
// 本文件定义了所有到达 common-service(:8080) 的请求如何分发：
//
//   ┌──────────────────────────────────────────────────┐
//   │              common-service :8080                 │
//   │          （API 网关 + 公共模块）                    │
//   │                                                  │
//   │  /api/user/*        → 本地处理（用户模块）         │
//   │  /api/address/*     → 本地处理（地址模块）         │
//   │  /api/expert/*      → 本地处理（专家模块）         │
//   │  /api/rural/*       → 本地处理（农村模块）         │
//   │  /api/upload/*      → 本地处理（上传模块）         │
//   │  ────────────────────────────────────────         │
//   │  /api/knowledge/*   → 代理 → knowledge :8081     │
//   │  /api/posts/*       → 代理 → post :8083          │
//   │  /api/products/*    → 代理 → product :8084       │
//   │  /api/orders/*      → 代理 → product :8084       │
//   │  /api/message/*     → 代理 → message :8082       │
//   │  /api/notification/*→ 代理 → message :8082       │
//   └──────────────────────────────────────────────────┘
//
// 设计原则：
//   ① 公共模块路由在前（/api/user, /api/address 等）
//      这些路由由 common-service 本地处理，不转发
//
//   ② 代理路由在后（/api/knowledge, /api/products 等）
//      这些路由通过 ProxyTo 透明转发到对应的微服务
//
//   ③ Gin 按注册顺序匹配路由，所以公共路由必须在前，
//      避免被通配符代理路由拦截
// ============================================================

// InitRouter 初始化 common-service 路由
// common-service 作为网关：先注册自身公共模块，未匹配的请求代理到对应微服务
func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 禁用尾斜杠/固定路径自动重定向
	// 原因：代理路由同时注册了 /knowledge/*path（通配）和 /knowledge（精确），
	// Gin 的 RedirectTrailingSlash 会在两者间无限 301 循环
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	// 静态文件服务：提供上传文件的访问，映射 /uploads 到本地 ./uploads 目录
	// 例如：http://localhost:8080/uploads/misc/xxx.jpg → ./uploads/misc/xxx.jpg
	r.Static("/uploads", "./uploads")

	api := r.Group("/api")
	{
		// ======================================================
		// 第一部分：公共模块路由（common-service 自身处理）
		// 这些是未拆分的保留模块，仍在 common-service 中运行
		// ======================================================
		UserRouter(api)           // 用户管理（登录/注册/个人信息）
		AddressRouter(api)        // 地址管理
		ExpertRouter(api)         // 专家管理
		RuralInfoRouter(api)      // 农村信息
		RuralAffairRouter(api)    // 农村事务
		AffairProcessorRouter(api) // 事务处理
		SensitiveWordRouter(api)  // 敏感词管理
		UploadRouter(api)         // 文件上传
		CommentRuralRouter(api)   // 农村评论

		// ======================================================
		// 第二部分：代理路由（转发到拆分后的微服务）
		// 每段代码的作用：
		//   1. ProxyTo("xxx-service") → 创建反向代理中间件
		//   2. api.Any("path", handler) → 注册路由（匹配 GET/POST/PUT/DELETE）
		//   3. /*path 通配符 → 匹配所有子路径
		// ======================================================

		// ── knowledge-service :8081 ── 知识库模块
		// 支持路径：/api/knowledge, /api/knowledge/123, /api/knowledges, /api/admin/knowledge/*, /api/comment_knowledge/* 等
		proxyKnowledge := middleware.ProxyTo("knowledge-service")
		api.Any("/knowledge/*path", proxyKnowledge)
		api.Any("/knowledge", proxyKnowledge)
		api.Any("/knowledges/*path", proxyKnowledge)
		api.Any("/knowledges", proxyKnowledge)
		api.Any("/admin/knowledge/*path", proxyKnowledge)
		api.Any("/comment_knowledge/*path", proxyKnowledge)

		// ── post-service :8083 ── 动态模块
		// 支持路径：/api/post, /api/post/123, /api/posts, /api/admin/post/* 等
		proxyPost := middleware.ProxyTo("post-service")
		api.Any("/post/*path", proxyPost)
		api.Any("/posts/*path", proxyPost)
		api.Any("/posts", proxyPost)
		api.Any("/admin/post/*path", proxyPost)

		// ── product-service :8084 ── 商品/订单/购物车模块
		// 一个微服务承载多个业务域：商品(product)、订单(order)、购物车(cart)、商品评论(comment_product)
		proxyProduct := middleware.ProxyTo("product-service")
		api.Any("/product/*path", proxyProduct)
		api.Any("/products/*path", proxyProduct)
		api.Any("/products", proxyProduct)
		api.Any("/admin/product/*path", proxyProduct)
		api.Any("/order/*path", proxyProduct)
		api.Any("/orders/*path", proxyProduct)
		api.Any("/orders", proxyProduct)
		api.Any("/cart/*path", proxyProduct)
		api.Any("/comment_product/*path", proxyProduct)

		// ── message-service :8082 ── 消息/通知模块
		proxyMessage := middleware.ProxyTo("message-service")
		api.Any("/message/*path", proxyMessage)
		api.Any("/notification/*path", proxyMessage)
		api.Any("/notification", proxyMessage)
	}

	return r
}
