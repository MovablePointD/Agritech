package config

import (
	"fmt"
	"log"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// ============================================================
// 角色：Nacos 注册中心客户端 — 对标 Spring Cloud Alibaba Nacos Discovery
//
// 本文件实现两个核心功能：
//   1. 服务注册   — 各微服务启动时向 Nacos 上报自己的 IP:Port
//   2. 服务注销   — 微服务关闭时从 Nacos 摘除自己
//
// 架构位置：
//   所有微服务模块（common/knowledge/message/post/product）启动时
//   均调用本文件，向同一个 Nacos Server 注册自身
//
//                        ┌─────────────────┐
//                        │  Nacos Server   │
//                        │ 127.0.0.1:8848  │
//                        └────────┬────────┘
//             注册 ┌──────────────┼──────────────┐ 注册
//         ┌───────┴───────┐                     ┌┴──────────────┐
//         │ common-service │  ... 四个微服务 ... │ product-service│
//         │    :8080       │                     │    :8084       │
//         └───────────────┘                     └───────────────┘
// ============================================================

// NacosConfig Nacos 连接配置
type NacosConfig struct {
	IpAddr      string // Nacos 服务地址
	Port        uint64 // Nacos 服务端口
	NamespaceId string // 命名空间ID（用于环境隔离：dev/test/prod）
	Group       string // 分组名称
	Username    string // 用户名（Nacos 2.2+ 默认开启鉴权）
	Password    string // 密码
	ContextPath string // 上下文路径（默认 /nacos）
}

// DefaultNacosConfig 默认Nacos配置（本地开发环境）
// 所有微服务启动时都会调用此函数获取统一的 Nacos 连接信息
func DefaultNacosConfig() NacosConfig {
	return NacosConfig{
		IpAddr:      "127.0.0.1",
		Port:        8848,
		NamespaceId: "13c09d73-0a85-40fc-b96c-b7d25a5d2821",
		Group:       "gocloud",
		Username:    "nacos",
		Password:    "nacos",
		ContextPath: "/nacos",
	}
}

// InitNamingClient 初始化 Nacos 命名客户端（服务注册与发现）
//
// 这是连接 Nacos 的第一步：创建一个可与 Nacos Server 通信的客户端。
// 参数说明：
//   - ServerConfig: Nacos Server 的地址、端口、上下文路径
//   - ClientConfig: 命名空间、鉴权信息、超时等客户端参数
func InitNamingClient(cfg NacosConfig) (naming_client.INamingClient, error) {
	ctxPath := cfg.ContextPath
	if ctxPath == "" {
		ctxPath = "/nacos"
	}

	// 配置 Nacos Server 连接信息
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(cfg.IpAddr, cfg.Port, constant.WithContextPath(ctxPath)),
	}

	// 配置 Nacos Client 客户端参数
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(cfg.NamespaceId), // 命名空间：用于环境隔离
		constant.WithUsername(cfg.Username),       // Nacos 鉴权用户名
		constant.WithPassword(cfg.Password),       // Nacos 鉴权密码
		constant.WithTimeoutMs(5000),              // 请求超时 5 秒
		constant.WithNotLoadCacheAtStart(true),    // 不从本地缓存加载（避免读到过期数据）
		constant.WithLogDir("/tmp/nacos/log"),     // Nacos SDK 日志目录
		constant.WithCacheDir("/tmp/nacos/cache"), // Nacos SDK 缓存目录
		constant.WithLogLevel("info"),
	)

	// 创建 NamingClient — 负责服务注册/发现的核心客户端
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("创建Nacos命名客户端失败: %w", err)
	}

	return client, nil
}

// RegisterService 向 Nacos 注册服务实例
//
// 对标 Spring Cloud 的 @EnableDiscoveryClient 自动注册行为。
// 每个微服务启动时调用此函数，将自己的 IP:Port 上报给 Nacos Server。
//
// 参数：
//   - client:   已初始化的 NamingClient
//   - serviceName: Nacos 中的服务名（如 "common-service", "knowledge-service"）
//   - ip:       本机 IP（传空则自动获取 127.0.0.1）
//   - port:     本服务监听的端口
func RegisterService(client naming_client.INamingClient, serviceName string, ip string, port uint64) error {
	if ip == "" {
		ip = getLocalIP()
	}

	// 向 Nacos 注册一个服务实例
	success, err := client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,              // 实例 IP
		Port:        port,            // 实例端口
		ServiceName: serviceName,     // 所属服务名
		GroupName:   "DEFAULT_GROUP", // 分组名（同一分组内的服务可互相发现）
		ClusterName: "DEFAULT",       // 集群名（同集群优先调用）
		Weight:      10,              // 权重（负载均衡用，数值越大分配流量越多）
		Enable:      true,            // 是否启用
		Healthy:     true,            // 是否健康
		Ephemeral:   true,            // 是否为临时实例（断连后自动剔除）
		Metadata: map[string]string{
			"version": "1.0.0", // 元数据：版本号等附加信息
		},
	})
	if err != nil {
		return fmt.Errorf("注册服务 %s 失败: %w", serviceName, err)
	}
	if !success {
		return fmt.Errorf("注册服务 %s 返回 false", serviceName)
	}

	log.Printf("[Nacos] 服务注册成功: %s@%s:%d", serviceName, ip, port)
	return nil
}

// DeregisterService 从 Nacos 注销服务实例
//
// 微服务优雅关闭时调用，从 Nacos 摘除自己的实例，
// 避免 Nacos 将已经下线的服务实例返回给调用方，导致请求失败。
// 通过 defer 注册，保证无论何种退出方式都会执行注销。
func DeregisterService(client naming_client.INamingClient, serviceName string, ip string, port uint64) {
	if ip == "" {
		ip = getLocalIP()
	}
	// 忽略注销错误（服务已关闭，即使注销失败也无法补救）
	_, _ = client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		GroupName:   "DEFAULT_GROUP",
		Cluster:     "DEFAULT", // 注意：Deregister 字段是 Cluster，不是 ClusterName
		Ephemeral:   true,
	})
	log.Printf("[Nacos] 服务注销成功: %s@%s:%d", serviceName, ip, port)
}

// getLocalIP 获取本机可用 IP（优先环回地址）
// 本地开发环境全部使用 127.0.0.1，避免虚拟网卡（如 Hyper-V、Docker）返回 10.x.x.x 等错误 IP
func getLocalIP() string {
	return "127.0.0.1"
}

// ParsePort 将端口字符串解析为 uint64
func ParsePort(port string) uint64 {
	p, err := strconv.ParseUint(port, 10, 64)
	if err != nil {
		return 8080 // 默认回退端口
	}
	return p
}
