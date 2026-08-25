# aiagent — AI 智能聊天服务

农技综合服务平台的 **AI 微服务**，基于 Flask 与 DeepSeek 大模型，提供农技问答、内容生成及与主平台联动发布能力。

---

## 项目简介

- **框架**：Flask 4.0 + Flask-CORS
- **大模型**：DeepSeek API（OpenAI 兼容接口）
- **会话存储**：MongoDB（不可用时自动降级为本地 JSON 文件）
- **默认端口**：5000

### 双助手模式

| 助手 | 功能 |
|------|------|
| **聊天助手** | 农业技术智能问答（种植、养殖、病虫害等） |
| **系统助手** | 文件辅助、内容生成，可联动发布动态/知识至主平台 |

与主后端通过 HTTP 集成：AI 服务调用 `http://localhost:8080` 网关接口发布内容，不侵入 Go 微服务代码。

---

## 环境要求

- **Python** 3.10（推荐使用 Conda 环境 `python310`）
- **DeepSeek API Key**（必填，[申请地址](https://platform.deepseek.com/)）
- **MongoDB** 4.4+（推荐；未安装时会使用 `contents/sessions_fallback.json` 降级）
- **后端 rxtcloud** 已启动（系统助手发布内容时需要）

---

## 启动步骤

### 1. 配置 API Key

复制或编辑项目根目录下的 `.env` 文件：

```env
# DeepSeek API Key（必填）
DEEPSEEK_API_KEY=你的_API_Key

# 服务配置
FLASK_HOST=0.0.0.0
FLASK_PORT=5000
FLASK_DEBUG=true

# 主后端地址（联动发布用）
BACKEND_URL=http://localhost:8080

# MongoDB（可选）
MONGO_URI=mongodb://localhost:27017
MONGO_DB=aiagent
```

### 2. 一键启动（Windows 推荐）

```bash
start.bat
```

脚本将自动：检查 Conda 环境 → 安装依赖 → 检查 `.env` → 启动服务。

### 3. 手动启动

```bash
# 创建并激活环境（示例）
conda create -n python310 python=3.10 -y
conda activate python310

pip install -r requirements.txt
python app.py
```

### 4. Linux / macOS

```bash
chmod +x start.sh
./start.sh
```

---

## 验证是否启动成功

- 健康检查：`http://localhost:5000/api/health`
- 前端访问：先启动 `rxtvue`，浏览器打开 `http://localhost:3000/ai-chat`

---

## 目录结构

```
aiagent/
├── app.py              # 主程序（Flask 应用）
├── requirements.txt    # Python 依赖
├── .env                # 环境变量（需自行配置 API Key）
├── start.bat / start.sh
├── prompts/            # 系统提示词
│   ├── chat_assistant.txt
│   └── system_assistant.txt
└── contents/           # 本地草稿与会话降级存储
```

---

## 主要 API

| 接口 | 说明 |
|------|------|
| `GET /api/health` | 服务健康检查 |
| `GET /api/sessions` | 会话列表 |
| `POST /api/chat/stream` | 流式对话（前端主要使用） |
| `POST /api/chat/publish` | AI 生成并发布内容 |
| `GET /api/models` | 可用助手列表 |

---

## 常见问题

**Q：提示 API Key 未配置？**  
A：编辑 `.env`，将 `DEEPSEEK_API_KEY` 替换为真实密钥。

**Q：聊天正常但无法发布到平台？**  
A：确认后端 `rxtcloud` 网关（8080）已启动，且用户已登录（需携带 JWT）。

**Q：MongoDB 未安装能否使用？**  
A：可以。服务会自动使用 `contents/sessions_fallback.json` 存储会话，功能略受限。

**Q：连接主服务失败？**  
A：检查 `.env` 中 `BACKEND_URL` 是否为 `http://localhost:8080`。
