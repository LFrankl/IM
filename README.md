# IM — Web Instant Messaging

仿 QQ 风格的 Web 即时通信系统，支持私聊、群聊、QQ 空间等核心功能。

## 技术栈

| 层 | 技术 |
|---|---|
| 前端 | Vue 3 + TypeScript + Pinia + Vue Router + Vite |
| 后端 | Go + Gin + GORM |
| 数据库 | SQLite |
| 实时通信 | WebSocket (gorilla/websocket) |
| 认证 | JWT |

## 功能模块

- **认证**：注册、登录、JWT 鉴权
- **好友系统**：搜索用户、发送/接受好友申请、删除好友、备注、分组
- **私聊**：实时消息、图片/文件发送、历史记录、未读计数
- **群聊**：创建群、搜索加入、退出/解散、踢人、实时群消息
- **QQ 空间**：发布动态（文字+图片）、点赞、评论、好友动态 Feed、个人主页

## UI 设计

仿 QQ NT（新技术）桌面端风格，三栏布局：

- 导航栏（68px）：深色背景 `#2C2C2C`，功能区入口
- 列表面板（280px）：浅灰背景 `#F5F5F5`，会话/联系人列表
- 内容区：主要交互区域

主色调为 `#1677FF`，气泡颜色区分：自己（`#C6E2FF`）/ 对方（`#FFFFFF`）。

## 项目结构

```
im/
├── backend/          # Go 后端
│   ├── main.go
│   ├── config/       # 配置文件
│   ├── internal/
│   │   ├── model/    # 数据模型
│   │   ├── dao/      # 数据访问层
│   │   ├── service/  # 业务逻辑
│   │   ├── handler/  # HTTP/WS 处理器
│   │   ├── middleware/
│   │   ├── router/
│   │   └── ws/       # WebSocket Hub
│   └── pkg/
│       ├── database/ # SQLite 初始化
│       ├── jwt/
│       └── response/ # 统一响应格式
└── frontend/         # Vue3 前端
    └── src/
        ├── api/      # 请求层
        ├── stores/   # Pinia 状态
        ├── views/    # 页面
        ├── components/
        └── composables/
```

## 快速启动

**后端**

```bash
cd backend
go run main.go
# 监听 :8080
```

**前端**

```bash
cd frontend
npm install
npm run dev
# 访问 http://localhost:5173
```

> 前端通过 Vite proxy 将 `/api` 和 `/ws` 请求转发到 `:8080`，无需额外跨域配置。

## WebSocket 协议

连接地址：`ws://localhost:8080/ws?token=<jwt>`

**客户端发送**

```json
{ "type": "chat_private", "to_id": 123, "content": "hello", "msg_type": "text" }
{ "type": "chat_group", "group_id": 456, "content": "hello", "msg_type": "text" }
```

**服务端推送**

```json
{ "type": "message", "data": { ... } }
{ "type": "friend_request", "data": { ... } }
{ "type": "message_sent", "data": { ... } }
```

## 配置

`backend/config/config.yaml`

```yaml
server:
  port: 8080
database:
  path: ./data/im.db
jwt:
  secret: "your-secret"
  expire: 168h
cors:
  allow_origins:
    - "http://localhost:5173"
```

## 数据存储

| 资源 | 路径 |
|------|------|
| 数据库文件 | `backend/data/im.db` |
| 上传文件 | `backend/data/uploads/` |
| 配置文件 | `backend/config/config.yaml` |

## 开发文档

- `docs/style-guide.md` — UI 设计规范（色彩、排版、组件）
- `docs/dev-progress.md` — 架构说明、数据库设计、阶段计划
