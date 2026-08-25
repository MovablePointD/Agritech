# rxtcloud — 后端微服务

农技综合服务平台的 **Go 微服务后端**，采用 API 网关 + 领域服务拆分架构，基于 Gin、GORM、Nacos 构建。

---

## 项目简介

`rxtcloud` 是平台的业务核心，包含 **1 个 API 网关服务** 与 **4 个领域微服务**：

| 服务 | 端口 | 职责 |
|------|------|------|
| **common-service** | 8080 | API 网关 + 用户/地址/专家/农村信息/农村事务/上传/敏感词 |
| **knowledge-service** | 8081 | 知识文章、知识评论、点赞 |
| **message-service** | 8082 | 私信、专家咨询、系统通知 |
| **post-service** | 8083 | 社区动态、动态评论、点赞 |
| **product-service** | 8084 | 商品、购物车、订单、商品评论 |

前端所有 `/api` 请求统一经 **8080 网关** 进入，由网关本地处理或反向代理至对应微服务。

---

## 环境要求

- **Go** 1.21+（工作区声明 1.25，低版本 Go 通常也可编译）
- **MySQL** 8.0+，数据库名 `gorxt_db`
- **Nacos** 2.x（可选，地址 `127.0.0.1:8848`）

默认数据库连接（`common/config/db.go`）：

```
root:root@tcp(127.0.0.1:3306)/gorxt_db
```

---

## 启动步骤

### 1. 准备数据库

在解压包根目录找到 `gorxt_db.sql`，导入 MySQL：

```bash
mysql -u root -p < ..\gorxt_db.sql
```

> 路径相对于 `rxtcloud` 目录，请按实际解压位置调整。

### 2. 启动 Nacos（推荐）

下载并启动 [Nacos](https://nacos.io/)，默认 `127.0.0.1:8848`。  
未启动 Nacos 时，服务仍可通过本地端口兜底运行。

### 3. 一键启动全部微服务

**Windows（推荐）：**

```bash
scripts\start_all.bat
```

将依次打开 5 个终端窗口，分别运行各服务。

**单独启动某个服务：**

```bash
scripts\start_common.bat      # 网关 :8080
scripts\start_knowledge.bat   # 知识 :8081
scripts\start_message.bat     # 消息 :8082
scripts\start_post.bat        # 动态 :8083
scripts\start_product.bat     # 电商 :8084
```

**手动启动（示例）：**

```bash
cd common
go run main.go
```

其余服务同理，进入对应目录执行 `go run main.go`。

### 4. 导入演示数据（可选）

```bash
scripts\seed_data.bat
```

---

## 验证是否启动成功

- 浏览器或 curl 访问：`http://localhost:8080/api/products`（应返回 JSON）
- 各服务控制台输出 `数据库连接成功`、路由注册信息
- 上传文件访问：`http://localhost:8080/uploads/...`

---

## 目录结构

```
rxtcloud/
├── common/              # 网关 + 公共业务模块
├── knowledge-service/   # 知识微服务
├── message-service/     # 消息微服务
├── post-service/        # 动态微服务
├── product-service/     # 电商微服务
├── scripts/             # 启动/停止/构建/种子数据脚本
├── docs/                # 架构图、数据流图等文档
└── go.work              # Go 工作区配置
```

---

## 停止服务

```bash
scripts\stop_all.bat
```

或直接关闭各服务对应的命令行窗口。

---

## 配置修改

| 配置项 | 位置 |
|--------|------|
| MySQL 连接 | 各服务 `config/db.go` |
| Nacos 地址 | `common/config/nacos.go` |
| JWT 密钥 | `common/utils/jwt.go`、`common/middleware/auth.go` |
| 服务端口 | 各服务 `main.go` / `config/config.go` |

修改数据库账号密码后，需同步更新所有微服务中的 DSN 字符串。

---

## 与前端、AI 的协作

- **前端 rxtvue**：通过 `http://localhost:8080` 访问 `/api` 与 `/uploads`
- **AI 服务 aiagent**：通过 `BACKEND_URL=http://localhost:8080` 调用网关发布动态/知识

请先启动本后端，再启动前端与 AI 服务。
