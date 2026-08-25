# rxtvue — 前端 Web 应用

农技综合服务平台的 **Vue 3 前端**，提供用户界面，涵盖农村事务、电商、知识社区、农村信息、AI 助手及管理后台等功能。

---

## 项目简介

- **框架**：Vue 3（Composition API）+ Vue Router 4
- **UI 库**：Element Plus（中文界面）
- **构建工具**：Vite 5
- **HTTP**：Axios，统一封装于 `src/utils/api.js`
- **富文本**：Quill + vue-quill

前端通过 Vite 开发代理将 `/api`、`/uploads` 转发至后端网关 `http://localhost:8080`；AI 聊天页面直连 `http://localhost:5000`。

---

## 环境要求

- **Node.js** 18 或更高版本
- **npm** 或 **pnpm** / **yarn**

后端（rxtcloud :8080）需已启动；使用 AI 助手时需同时启动 aiagent（:5000）。

---

## 启动步骤

### 1. 安装依赖

```bash
cd rxtvue
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

启动成功后终端会显示本地访问地址，默认：

**http://localhost:3000**

在浏览器中打开即可使用。

### 3. 生产构建（可选）

```bash
npm run build
npm run preview
```

构建产物输出至 `dist/` 目录。

---

## 主要页面

| 路径 | 功能 |
|------|------|
| `/` | 首页 |
| `/login` `/register` | 登录注册 |
| `/products` `/cart` `/orders` | 电商 |
| `/knowledge` `/posts` | 知识与动态 |
| `/affairs` `/processor` | 农村事务与处理中心 |
| `/rural-info` | 农村信息中心 |
| `/ai-chat` | AI 智能助手 |
| `/user` | 个人中心 |
| `/admin` `/reviewer` | 系统管理与审核 |

---

## 配置说明

代理配置位于 `vite.config.js`：

```javascript
proxy: {
  '/api': 'http://localhost:8080',
  '/uploads': 'http://localhost:8080'
}
```

若后端部署在其他地址，请修改 `target` 为实际网关 URL。

AI 服务地址在 `src/views/AIChat.vue` 中配置为 `http://localhost:5000/api`，部署时需改为实际 AI 服务地址。

---

## 目录结构

```
rxtvue/
├── src/
│   ├── main.js           # 入口
│   ├── router/           # 路由
│   ├── utils/api.js      # API 封装
│   ├── components/       # 公共组件
│   └── views/            # 页面组件
├── index.html
├── vite.config.js
└── package.json
```

---

## 常见问题

**Q：页面能打开但接口报错？**  
A：确认后端 `rxtcloud` 已启动，且 `http://localhost:8080` 可访问。

**Q：登录后跳转异常？**  
A：清除浏览器 `localStorage` 中的 `token`、`user` 后重新登录。

**Q：AI 助手提示服务不可用？**  
A：确认 `aiagent` 已在 5000 端口运行，且 DeepSeek API Key 已配置。

---

## 相关项目

- 后端：同级目录 `rxtcloud/`
- AI 服务：同级目录 `aiagent/`
- 总体说明：上级目录 `README.md`
