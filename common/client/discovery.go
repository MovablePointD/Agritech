package client

import (
	"fmt"
	"log"
	"math/rand"
	"sync"

	"rxtcloud/common/config"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// ============================================================
// 角色：服务发现客户端 — 对标 Spring Cloud Nacos Discovery 的发现功能
//
// 本文件解决的核心问题：
//   调用方不需要硬编码目标服务的 IP:Port，只需知道目标服务名，
//   由 ServiceDiscovery 自动向 Nacos 查询对应服务在哪个 IP 和端口。
//
// 调用链路示例（浏览器请求 /api/products）：
//   common-service(:8080) → ProxyTo("product-service")
//        → ServiceDiscovery.GetServiceURL("product-service")
//        → Nacos 查询 product-service 的健康实例
//        → 返回 http://127.0.0.1:8084
//        → 反向代理转发到 product-service
//
// 容错机制（双保险）：
//   ① Nacos 可用 → 通过 Nacos 查询真实 IP:Port（可动态扩缩容）
//   ② Nacos 不可用 → 本地回退 fallbackURL() 写死 127.0.0.1:808x
//   这样即使 Nacos 宕机，本地开发和测试也不会受影响
// ============================================================

// ServiceDiscovery 基于 Nacos 的服务发现
type ServiceDiscovery struct {
	client naming_client.INamingClient // Nacos 命名客户端，用于查询服务实例
	mu     sync.RWMutex               // 读写锁，保护并发访问
}

var (
	discovery     *ServiceDiscovery
	discoveryOnce sync.Once // 单例模式，全局只创建一个 ServiceDiscovery
)

// GetDiscovery 获取服务发现单例（懒加载 + 线程安全）
//
// 为什么用单例？
//   Nacos 客户端需要维护与服务端的心跳连接，创建多个客户端会浪费资源。
//   全局共享一个实例，所有 ProxyTo 和 CallService 都通过它发现服务。
func GetDiscovery() *ServiceDiscovery {
	discoveryOnce.Do(func() {
		nacosCfg := config.DefaultNacosConfig()
		client, err := config.InitNamingClient(nacosCfg)
		if err != nil {
			// Nacos 初始化失败也不 panic，服务仍可通过 fallbackURL 运行
			log.Printf("[Discovery] Nacos 初始化失败，服务发现不可用: %v", err)
			discovery = &ServiceDiscovery{} // client 为 nil，后续所有查询走 fallbackURL
			return
		}
		discovery = &ServiceDiscovery{client: client}
		log.Println("[Discovery] 服务发现初始化成功")
	})
	return discovery
}

// GetServiceURL 根据服务名获取一个健康实例的 URL
//
// 这是服务发现最核心的方法，被 ProxyTo 和 CallService 共同调用。
//
// 流程：
//  1. 如果 Nacos 客户端为 nil → 走本地回退
//  2. 向 Nacos 查询 serviceName 的所有健康实例
//  3. 随机挑一个（简单负载均衡）
//  4. 拼接成 http://ip:port 返回
//
// 返回格式: http://ip:port
func (d *ServiceDiscovery) GetServiceURL(serviceName string) (string, error) {
	// 情况①：Nacos 客户端初始化失败，直接走本地回退
	if d.client == nil {
		return fallbackURL(serviceName)
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	// 情况②：Nacos 可用，查询健康实例
	instances, err := d.client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   "DEFAULT_GROUP",          // 只查同一分组的服务
		Clusters:    []string{"DEFAULT"},       // 只查同一集群的服务
		HealthyOnly: true,                      // 只返回健康的实例
	})
	if err != nil || len(instances) == 0 {
		// 情况③：Nacos 查询失败或没有实例，走本地回退
		log.Printf("[Discovery] Nacos 未找到 %s，尝试本地回退", serviceName)
		return fallbackURL(serviceName)
	}

	// 简单随机负载均衡：在所有健康实例中随机选一个
	instance := instances[rand.Intn(len(instances))]
	url := fmt.Sprintf("http://%s:%d", instance.Ip, instance.Port)
	return url, nil
}

// GetServiceURLs 获取服务的所有健康实例 URL（批量发现，用于自定义负载均衡场景）
func (d *ServiceDiscovery) GetServiceURLs(serviceName string) ([]string, error) {
	if d.client == nil {
		url, err := fallbackURL(serviceName)
		if err != nil {
			return nil, err
		}
		return []string{url}, nil
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	instances, err := d.client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   "DEFAULT_GROUP",
		Clusters:    []string{"DEFAULT"},
		HealthyOnly: true,
	})
	if err != nil {
		return nil, err
	}

	urls := make([]string, 0, len(instances))
	for _, inst := range instances {
		urls = append(urls, fmt.Sprintf("http://%s:%d", inst.Ip, inst.Port))
	}
	return urls, nil
}

// fallbackURL Nacos 不可用时的本地回退（开发环境）
//
// 硬编码每个微服务在本地开发时的默认端口映射。
// 当 Nacos 不可用时，请求仍能正确路由到本机对应端口的微服务。
//
// 端口分配表（避免冲突，每个服务独占一个端口）：
//   common-service     → 8080（API 网关 + 公共模块）
//   knowledge-service  → 8081（知识库模块）
//   message-service    → 8082（消息/通知模块）
//   post-service       → 8083（动态模块）
//   product-service    → 8084（商品/订单模块）
func fallbackURL(serviceName string) (string, error) {
	ports := map[string]string{
		"common-service":    "8080",
		"knowledge-service": "8081",
		"message-service":   "8082",
		"post-service":      "8083",
		"product-service":   "8084",
	}
	if port, ok := ports[serviceName]; ok {
		log.Printf("[Discovery] 回退到本地 %s -> 127.0.0.1:%s", serviceName, port)
		return fmt.Sprintf("http://127.0.0.1:%s", port), nil
	}
	return "", fmt.Errorf("未找到服务 %s 的实例（Nacos不可用且无本地映射）", serviceName)
}
