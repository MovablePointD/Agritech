package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// ============================================================
// 角色：服务间 HTTP 调用客户端 — 对标 Spring Cloud OpenFeign
//
// 本文件是微服务间通信的核心执行引擎。
//
// OpenFeign 在 Java 中做的事情：
//   1. 声明接口 (@FeignClient) → 自动生成 HTTP 客户端
//   2. 集成 Ribbon/LoadBalancer → 自动负载均衡
//   3. 集成 Nacos Discovery → 自动服务发现
//
// 本文件在 Go 中的对应实现：
//   1. CallService() → 服务名 + API路径 → 自动发现 + HTTP调用（Feign 的核心功能）
//   2. CallAsync()   → fire-and-forget 异步调用（Feign 没有直接对应，是增强功能）
//   3. discovery      → 内嵌 Nacos 服务发现（Feign + Nacos 的整合）
//
// 架构位置：
//   message-service 需要转发点赞给 post-service 时，不是直接 import post-service
//   的代码，而是调用本客户端的 CallService("post-service", "POST", "/api/internal/like", body, nil)
//
//                  ┌──────────────────┐
//                  │ message-service  │
//                  │   (调用方)        │
//                  └────────┬─────────┘
//                           │ ServiceClient.CallService("post-service", ...)
//                           ▼
//                  ┌──────────────────┐
//                  │  ServiceClient   │ ← 本文件
//                  │  (OpenFeign角色)  │
//                  └────────┬─────────┘
//                           │ discovery.GetServiceURL("post-service")
//                           ▼
//                  ┌──────────────────┐
//                  │  Nacos Server    │
//                  └────────┬─────────┘
//                           │ 返回 127.0.0.1:8083
//                           ▼
//                  ┌──────────────────┐
//                  │  post-service    │
//                  │  (被调用方)       │
//                  └──────────────────┘
// ============================================================

// ServiceClient 服务间 HTTP 调用客户端
// 封装了 "发现目标服务地址 → 组装 HTTP 请求 → 发送 → 解析响应" 的完整流程
type ServiceClient struct {
	httpClient  *http.Client       // Go 标准 HTTP 客户端（可配置超时、连接池等）
	discovery   *ServiceDiscovery  // Nacos 服务发现（获取目标服务地址）
	serviceName string             // 调用方自身的服务名（写入 X-Service-From 请求头）
}

// NewServiceClient 创建服务调用客户端
func NewServiceClient(serviceName string) *ServiceClient {
	return &ServiceClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // 10 秒超时，避免调用方被下游拖死
		},
		discovery:   GetDiscovery(),
		serviceName: serviceName,
	}
}

// CallService 通过 Nacos 服务发现调用目标服务（核心方法，对标 OpenFeign）
//
// 参数：
//   targetService: 目标服务名（如 "common-service", "knowledge-service"）
//   method:       HTTP 方法（GET/POST/PUT/DELETE）
//   path:         API 路径（如 "/api/user/info/1", 前面不带主机名）
//   body:         请求体（结构体会自动 JSON 序列化），无 body 传 nil
//   result:       响应体（JSON 反序列化到此结构体），不关心响应传 nil
//
// 完整调用链路：
//   CallService("common-service", "GET", "/api/user/info/1", nil, &userInfo)
//     → discovery.GetServiceURL("common-service")     // Nacos 发现: http://127.0.0.1:8080
//     → fullURL = "http://127.0.0.1:8080/api/user/info/1"
//     → doRequest("GET", fullURL, nil, &userInfo)     // HTTP 调用 + JSON 解析
func (c *ServiceClient) CallService(targetService, method, path string, body interface{}, result interface{}) error {
	// ① 通过 Nacos 发现目标服务地址（或走本地回退）
	baseURL, err := c.discovery.GetServiceURL(targetService)
	if err != nil {
		return fmt.Errorf("服务发现失败 [%s]: %w", targetService, err)
	}

	// ② 拼接完整 URL
	fullURL := baseURL + path
	return c.doRequest(method, fullURL, body, result)
}

// CallServiceWithURL 直接使用完整 URL 调用（跳过服务发现，直接 HTTP）
// 用于调用非 Nacos 注册的外部服务，或者已经知道确切地址的场景
func (c *ServiceClient) CallServiceWithURL(url, method string, body interface{}, result interface{}) error {
	return c.doRequest(method, url, body, result)
}

// doRequest 执行 HTTP 请求的内部方法
//
// 流程：
//   1. 如果 body 不为 nil → JSON 序列化
//   2. 组装 http.Request，设置 Content-Type 和 X-Service-From 头
//   3. 发送请求
//   4. 检查 HTTP 状态码（非 2xx 视为错误）
//   5. 如果 result 不为 nil → JSON 反序列化响应体
func (c *ServiceClient) doRequest(method, url string, body interface{}, result interface{}) error {
	// ① 序列化请求体
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求体失败: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	// ② 创建 HTTP 请求
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// ③ 设置请求头
	req.Header.Set("Content-Type", "application/json")
	// X-Service-From：标识调用来源，被调用方可据此做日志追踪或权限控制
	req.Header.Set("X-Service-From", c.serviceName)

	// ④ 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求 %s %s 失败: %w", method, url, err)
	}
	defer resp.Body.Close()

	// ⑤ 状态码校验（非 2xx 统一视为调用失败）
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("服务返回错误 [%d] %s: %s", resp.StatusCode, url, string(respBody))
	}

	// ⑥ 反序列化响应（调用方传入了 result 才解析）
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}

	return nil
}

// CallAsync 异步调用目标服务（fire-and-forget）
//
// 用于无需等待返回结果的场景，例如：
//   - 点赞操作转发到其他微服务（同步给 knowledge-service 即可，不用等返回）
//   - 发送通知
//   - 记录操作日志
//
// 内部启动 goroutine 异步执行，调用方不会被阻塞。
// 失败只打印日志，不向上传递错误（fire-and-forget 语义）。
func (c *ServiceClient) CallAsync(targetService, method, path string, body interface{}) {
	go func() {
		if err := c.CallService(targetService, method, path, body, nil); err != nil {
			log.Printf("[Async] 异步调用 %s%s 失败: %v", targetService, path, err)
		}
	}()
}
