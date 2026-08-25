package client

import (
	"fmt"
	"log"
)

// ============================================================
// 角色：服务桥接层 — 对标 Spring Cloud OpenFeign 的 @FeignClient 接口定义
//
// 如果说 http_client.go 是 Feign 的底层执行引擎，
// 那么 service_bridge.go 就是 Feign 的接口层：
//
//   Java OpenFeign 写法：
//     @FeignClient(name = "post-service")
//     interface PostServiceClient {
//         @PostMapping("/api/internal/like")
//         LikeResult forwardLike(LikeBody body);
//     }
//
//   本文件 Go 对应：
//     func ForwardLike(body LikeCallbackBody, callerName string) (*LikeCallbackResult, error)
//
// 本文件封装了三类跨服务调用接口：
//   1. ForwardLike          — 点赞同步（message → post/knowledge/product）
//   2. GetUserInfo          — 获取用户信息（子服务 → common-service）
//   3. CheckContent         — 敏感词检测（子服务 → common-service）
//
// 跨服务调用拓扑图：
//
//             ┌─────────────┐
//             │   common    │ ← 用户信息 / 敏感词检测
//             │  :8080      │
//             └──▲──▲──▲────┘
//                │  │  │  GetUserInfo() / CheckContent()
//   ┌────────────┼──┼──┼───────────────┐
//   │            │  │  │               │
//   │   ┌────────┴──┴──┴─────────┐     │
//   │   │     message-service    │     │
//   │   │        :8082           │─────┤  ForwardLike() →
//   │   └────────────────────────┘     │    同步点赞到
//   │                                  │    post / knowledge / product
//   │   ┌──────────────┐  ┌───────────┐│
//   │   │post-service  │  │knowledge  ││
//   │   │   :8083      │  │  :8081    ││
//   │   └──────────────┘  └───────────┘│
//   │                                  │
//   │   ┌──────────────┐               │
//   │   │product-svc   │               │
//   │   │   :8084      │               │
//   └───┴──▲──▲──▲─────┴───────────────┘
//          │  │  │
//     ForwardLike 转发到达三家
// ============================================================

// ============================================================
// 一、点赞转发
// ============================================================

// LikeCallbackBody Like 回调请求体
// 由 message-service 组装，发送给 post/knowledge/product-service
type LikeCallbackBody struct {
	UserID     uint   `json:"user_id"`     // 点赞用户 ID
	TargetType string `json:"target_type"` // 点赞对象类型: post / post_comment / knowledge / knowledge_comment / product_comment
	TargetID   uint   `json:"target_id"`   // 点赞对象 ID
	IsLike     bool   `json:"is_like"`     // true=点赞, false=取消点赞
}

// LikeCallbackResult Like 回调结果
type LikeCallbackResult struct {
	Liked bool `json:"liked"` // 当前是否已点赞
	Likes int  `json:"likes"` // 最新点赞总数
}

// targetTypeToService 将 Like 的 targetType 映射到目标服务名和 API 路径
//
// 映射规则：
//   post / post_comment         → post-service     /api/internal/like
//   knowledge / knowledge_comment → knowledge-service /api/internal/like
//   product_comment             → product-service  /api/internal/like
//
// 每个目标服务都暴露了 /api/internal/like 内部接口来接收点赞同步。
// "internal" 前缀表示这些接口仅供微服务间内部调用，不应暴露给前端。
func targetTypeToService(targetType string) (serviceName, path string) {
	switch targetType {
	case "post", "post_comment":
		return "post-service", "/api/internal/like"
	case "knowledge", "knowledge_comment":
		return "knowledge-service", "/api/internal/like"
	case "product_comment":
		return "product-service", "/api/internal/like"
	default:
		return "", ""
	}
}

// ForwardLike 将点赞操作转发到目标服务（同步调用，需要返回结果）
//
// 由 message-service 调用，将点赞操作同步到对应的业务服务。
// 调用方可以从返回值中获取最新的点赞状态和点赞数。
//
// 使用场景：
//   message-service 中的点赞接口收到用户点赞请求后：
//   ① 先写入 message 自己的点赞表
//   ② 再调用 ForwardLike 同步到目标服务（如 post-service）
//   ③ post-service 更新自己表的 like_count，返回最新状态
func ForwardLike(body LikeCallbackBody, callerServiceName string) (*LikeCallbackResult, error) {
	targetService, path := targetTypeToService(body.TargetType)
	if targetService == "" {
		return nil, fmt.Errorf("未知的点赞目标类型: %s", body.TargetType)
	}

	client := NewServiceClient(callerServiceName)
	var result LikeCallbackResult
	if err := client.CallService(targetService, "POST", path, body, &result); err != nil {
		return nil, fmt.Errorf("转发点赞到 %s 失败: %w", targetService, err)
	}
	return &result, nil
}

// ForwardLikeAsync 异步转发点赞操作（fire-and-forget，不关心返回）
//
// 适用场景：点赞只需要通知目标服务更新，调用方不需要等待结果。
// 内部启动 goroutine，立即返回，不阻塞调用方。
func ForwardLikeAsync(body LikeCallbackBody, callerServiceName string) {
	targetService, path := targetTypeToService(body.TargetType)
	if targetService == "" {
		log.Printf("[Like] 未知的点赞目标类型: %s", body.TargetType)
		return
	}

	client := NewServiceClient(callerServiceName)
	client.CallAsync(targetService, "POST", path, body)
}

// ============================================================
// 二、用户信息查询（跨服务数据共享）
// ============================================================

// UserInfoResult 用户信息响应
// 所有微服务模块需要展示用户信息时，都通过此结构从 common-service 获取
type UserInfoResult struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar_url"`
	Role     string `json:"role"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

// GetUserInfo 从 common-service 获取用户信息
//
// common-service 是唯一管理用户数据的服务。
// 其他微服务（如 knowledge-service 展示知识作者信息时）
// 不直接访问 users 表，而是通过此函数远程调用 common-service。
//
// 这样做的好处：
//   ① 数据主权归 common-service（避免多服务同时读写同一张表）
//   ② 如果将来拆出独立的 user-service，只需修改配置，调用方代码不变
//
// 使用示例（在 knowledge-service 中）：
//   user, err := client.GetUserInfo(knowledge.AuthorID, "knowledge-service")
//   if err != nil { /* 降级：只显示 ID，不显示昵称 */ }
func GetUserInfo(userID uint, callerServiceName string) (*UserInfoResult, error) {
	client := NewServiceClient(callerServiceName)
	var result UserInfoResult
	path := fmt.Sprintf("/api/user/info/%d", userID)
	if err := client.CallService("common-service", "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	return &result, nil
}

// ============================================================
// 三、敏感词检测（跨服务内容审核）
// ============================================================

// CheckContentReq 敏感词检测请求
type CheckContentReq struct {
	Content string `json:"content"` // 待检测的文本内容
}

// CheckContentResp 敏感词检测响应
type CheckContentResp struct {
	Valid     bool     `json:"valid"`              // 是否通过检测
	BadWords  []string `json:"bad_words,omitempty"` // 命中的敏感词列表
	CleanText string   `json:"clean_text,omitempty"` // 清洗后的文本
}

// CheckContent 调用 common-service 检测敏感词
//
// common-service 维护全局敏感词库，并提供检测接口。
// 各业务服务在发布/评论内容前，通过此函数进行审核。
//
// 使用示例（在 knowledge-service 中）：
//   resp, err := client.CheckContent(content, "knowledge-service")
//   if err != nil || !resp.Valid {
//       return errors.New("内容包含敏感词: " + strings.Join(resp.BadWords, ","))
//   }
func CheckContent(content string, callerServiceName string) (*CheckContentResp, error) {
	client := NewServiceClient(callerServiceName)
	var result CheckContentResp
	if err := client.CallService("common-service", "POST", "/api/check-content",
		CheckContentReq{Content: content}, &result); err != nil {
		return nil, fmt.Errorf("敏感词检测失败: %w", err)
	}
	return &result, nil
}
