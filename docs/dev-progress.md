# 即时通信系统 - 开发进度文档

> QQ 风格 Web 即时通信系统
> 技术栈：Vue3 + Go + SQLite
> 状态：准备阶段

---

## 项目概览

### 目标功能

| 模块 | 功能清单 |
|------|---------|
| **认证** | 注册、登录、登出、记住登录 |
| **好友系统** | 好友申请/同意/拒绝、删除好友、好友列表、分组（默认/自定义）、好友备注、查看资料 |
| **私聊** | 文字消息、图片消息、文件传输（本地路径）、消息历史、未读计数、已读回执 |
| **群组** | 创建群、搜索加入、退出群、群主踢人、解散群、群消息、查看全部群成员 |
| **QQ 空间** | 发布动态（文字+图片）、评论、点赞、个人主页（动态列表+封面） |
| **基础 UI** | 三栏布局、导航栏、搜索、右键菜单、通知提示、在线状态 |

### 不在范围内（本期）
- 语音/视频通话
- 消息撤回
- @功能
- 好友权限控制（空间仅自己可见等）
- 群管理员角色
- 文件夹式空间相册
- 移动端适配

---

## 技术架构

### 前端

```
技术选型：
  框架：      Vue 3.4+ (Composition API, <script setup>)
  构建工具：  Vite 5.x
  语言：      TypeScript 5.x
  路由：      Vue Router 4.x
  状态管理：  Pinia 2.x
  HTTP：      axios
  WebSocket： 原生 WebSocket（封装为 composable）
  UI 组件：   自研（仿 QQ 风格，无重型 UI 库依赖）
  图标：      @iconify/vue（mdi 图标集）
  样式：      CSS Variables + Scoped CSS（不使用 Tailwind）
```

### 后端

```
技术选型：
  语言：    Go 1.21+
  框架：    Gin
  ORM：     GORM v2
  数据库：  SQLite（mattn/go-sqlite3）
  实时通信：gorilla/websocket
  认证：    JWT（golang-jwt/jwt v5）
  配置：    viper
  文件存储：本地文件系统（上传目录）
```

### 数据存储

```
SQLite 文件位置：./backend/data/im.db
上传文件位置：  ./backend/data/uploads/
配置文件位置：  ./backend/config/config.yaml
```

---

## 项目目录结构

### 完整骨架

```
im/
├── docs/
│   ├── style-guide.md        # UI 设计规范（本项目参考）
│   └── dev-progress.md       # 本文件
│
├── frontend/
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── package.json
│   └── src/
│       ├── main.ts
│       ├── App.vue
│       │
│       ├── assets/
│       │   ├── styles/
│       │   │   ├── variables.css     # CSS 变量（来自 style-guide.md）
│       │   │   ├── reset.css         # 样式重置
│       │   │   └── global.css        # 全局基础样式
│       │   └── icons/                # 自定义 SVG 图标
│       │
│       ├── types/                    # TypeScript 类型定义
│       │   ├── user.ts
│       │   ├── chat.ts
│       │   ├── group.ts
│       │   └── space.ts
│       │
│       ├── api/                      # API 请求层
│       │   ├── client.ts             # axios 实例 + 拦截器
│       │   ├── auth.ts
│       │   ├── user.ts
│       │   ├── friend.ts
│       │   ├── chat.ts
│       │   ├── group.ts
│       │   └── space.ts
│       │
│       ├── stores/                   # Pinia stores
│       │   ├── auth.ts               # 认证状态（当前用户、token）
│       │   ├── chat.ts               # 会话列表、消息缓存
│       │   ├── contacts.ts           # 好友列表
│       │   ├── groups.ts             # 群组列表
│       │   └── space.ts              # 空间动态
│       │
│       ├── composables/              # 可复用逻辑
│       │   ├── useWebSocket.ts       # WebSocket 连接管理
│       │   ├── useMessages.ts        # 消息发送/接收逻辑
│       │   ├── useOnlineStatus.ts    # 在线状态
│       │   └── useContextMenu.ts    # 右键菜单
│       │
│       ├── router/
│       │   └── index.ts             # 路由配置（守卫）
│       │
│       ├── components/
│       │   ├── layout/
│       │   │   ├── AppLayout.vue        # 整体三栏布局容器
│       │   │   ├── NavBar.vue           # 左侧导航栏（68px）
│       │   │   ├── ListPanel.vue        # 中间列表栏（280px）
│       │   │   └── ContentArea.vue      # 右侧内容区
│       │   │
│       │   ├── common/
│       │   │   ├── Avatar.vue           # 头像组件（支持状态角标）
│       │   │   ├── Badge.vue            # 未读角标
│       │   │   ├── SearchBar.vue        # 搜索框
│       │   │   ├── ContextMenu.vue      # 右键菜单（teleport to body）
│       │   │   ├── Modal.vue            # 通用模态框
│       │   │   ├── Toast.vue            # 全局 Toast 提示
│       │   │   ├── Dropdown.vue         # 下拉组件
│       │   │   └── EmptyState.vue       # 空状态占位
│       │   │
│       │   ├── chat/
│       │   │   ├── ChatWindow.vue        # 聊天窗口容器（私聊/群聊）
│       │   │   ├── ChatHeader.vue        # 聊天顶部栏
│       │   │   ├── MessageList.vue       # 消息列表（虚拟滚动）
│       │   │   ├── MessageItem.vue       # 单条消息（含气泡）
│       │   │   ├── ChatBubble.vue        # 气泡组件（文字/图片/文件）
│       │   │   ├── ChatInput.vue         # 输入框区域
│       │   │   └── ChatToolbar.vue       # 输入工具栏（emoji/图片/文件）
│       │   │
│       │   ├── contacts/
│       │   │   ├── ContactList.vue       # 联系人列表（带分组折叠）
│       │   │   ├── ContactItem.vue       # 联系人列表项
│       │   │   ├── FriendRequest.vue     # 好友申请列表
│       │   │   └── UserProfile.vue       # 用户资料卡片（侧边栏或弹窗）
│       │   │
│       │   ├── groups/
│       │   │   ├── GroupList.vue         # 群组列表
│       │   │   ├── GroupItem.vue         # 群组列表项
│       │   │   ├── GroupMemberList.vue   # 群成员列表（侧边栏）
│       │   │   └── CreateGroupModal.vue  # 创建群弹窗
│       │   │
│       │   └── space/
│       │       ├── SpaceFeed.vue         # 动态列表
│       │       ├── SpacePost.vue         # 单条动态卡片
│       │       ├── SpaceEditor.vue       # 发布动态编辑器
│       │       ├── CommentList.vue       # 评论列表
│       │       └── ProfilePage.vue       # 个人主页
│       │
│       └── views/
│           ├── LoginView.vue             # 登录/注册页面
│           ├── MainView.vue              # 主界面（三栏布局）
│           └── SpaceView.vue            # QQ 空间独立视图
│
├── backend/
│   ├── go.mod
│   ├── go.sum
│   ├── main.go                     # 入口
│   │
│   ├── config/
│   │   └── config.yaml             # 配置文件
│   │
│   ├── data/                       # 运行时数据（gitignore）
│   │   ├── im.db                   # SQLite 数据库
│   │   └── uploads/                # 上传文件
│   │
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go           # 配置结构体 + 加载
│   │   │
│   │   ├── model/                  # 数据模型（GORM）
│   │   │   ├── user.go
│   │   │   ├── friend.go
│   │   │   ├── message.go
│   │   │   ├── group.go
│   │   │   └── space.go
│   │   │
│   │   ├── dao/                    # 数据访问层
│   │   │   ├── user.go
│   │   │   ├── friend.go
│   │   │   ├── message.go
│   │   │   ├── group.go
│   │   │   └── space.go
│   │   │
│   │   ├── service/                # 业务逻辑层
│   │   │   ├── auth.go
│   │   │   ├── friend.go
│   │   │   ├── chat.go
│   │   │   ├── group.go
│   │   │   └── space.go
│   │   │
│   │   ├── handler/                # HTTP/WS 处理器
│   │   │   ├── auth.go
│   │   │   ├── user.go
│   │   │   ├── friend.go
│   │   │   ├── chat.go
│   │   │   ├── group.go
│   │   │   ├── space.go
│   │   │   └── websocket.go
│   │   │
│   │   ├── middleware/
│   │   │   ├── auth.go             # JWT 验证中间件
│   │   │   ├── cors.go             # CORS 配置
│   │   │   └── logger.go           # 请求日志
│   │   │
│   │   ├── ws/                     # WebSocket 管理
│   │   │   ├── hub.go              # 连接中心（广播/路由）
│   │   │   └── client.go           # 单连接
│   │   │
│   │   └── router/
│   │       └── router.go           # 路由注册
│   │
│   └── pkg/
│       ├── database/
│       │   └── sqlite.go           # SQLite 初始化 + 迁移
│       ├── jwt/
│       │   └── jwt.go              # JWT 工具
│       └── response/
│           └── response.go         # 统一响应格式
```

---

## 数据库设计

### users 表

```sql
CREATE TABLE users (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  username    TEXT NOT NULL UNIQUE,
  password    TEXT NOT NULL,        -- bcrypt hash
  nickname    TEXT NOT NULL,
  avatar      TEXT DEFAULT '',      -- 头像路径
  bio         TEXT DEFAULT '',      -- 个性签名
  status      TEXT DEFAULT 'offline', -- online/offline/busy
  created_at  DATETIME NOT NULL,
  updated_at  DATETIME NOT NULL
);
```

### friendships 表

```sql
CREATE TABLE friendships (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id     INTEGER NOT NULL,
  friend_id   INTEGER NOT NULL,
  remark      TEXT DEFAULT '',      -- 好友备注
  group_name  TEXT DEFAULT '我的好友',
  created_at  DATETIME NOT NULL,
  UNIQUE(user_id, friend_id),
  FOREIGN KEY(user_id) REFERENCES users(id),
  FOREIGN KEY(friend_id) REFERENCES users(id)
);
```

### friend_requests 表

```sql
CREATE TABLE friend_requests (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  from_id     INTEGER NOT NULL,
  to_id       INTEGER NOT NULL,
  message     TEXT DEFAULT '',      -- 验证消息
  status      TEXT DEFAULT 'pending', -- pending/accepted/rejected
  created_at  DATETIME NOT NULL,
  updated_at  DATETIME NOT NULL,
  FOREIGN KEY(from_id) REFERENCES users(id),
  FOREIGN KEY(to_id) REFERENCES users(id)
);
```

### messages 表

```sql
CREATE TABLE messages (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  from_id     INTEGER NOT NULL,
  to_id       INTEGER NOT NULL,    -- 私聊: 对方 user_id; 群聊: group_id（负数区分）
  type        TEXT NOT NULL,       -- text/image/file
  content     TEXT NOT NULL,       -- 消息内容（JSON）
  chat_type   TEXT NOT NULL,       -- private/group
  is_read     INTEGER DEFAULT 0,
  created_at  DATETIME NOT NULL,
  FOREIGN KEY(from_id) REFERENCES users(id)
);
CREATE INDEX idx_messages_chat ON messages(chat_type, to_id, created_at);
```

### groups 表

```sql
CREATE TABLE groups (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  name        TEXT NOT NULL,
  avatar      TEXT DEFAULT '',
  notice      TEXT DEFAULT '',     -- 群公告
  owner_id    INTEGER NOT NULL,    -- 群主
  created_at  DATETIME NOT NULL,
  updated_at  DATETIME NOT NULL,
  FOREIGN KEY(owner_id) REFERENCES users(id)
);
```

### group_members 表

```sql
CREATE TABLE group_members (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  group_id    INTEGER NOT NULL,
  user_id     INTEGER NOT NULL,
  joined_at   DATETIME NOT NULL,
  UNIQUE(group_id, user_id),
  FOREIGN KEY(group_id) REFERENCES groups(id),
  FOREIGN KEY(user_id) REFERENCES users(id)
);
```

### space_posts 表

```sql
CREATE TABLE space_posts (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id     INTEGER NOT NULL,
  content     TEXT NOT NULL,
  images      TEXT DEFAULT '[]',   -- JSON 数组，图片路径列表
  like_count  INTEGER DEFAULT 0,
  created_at  DATETIME NOT NULL,
  updated_at  DATETIME NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id)
);
```

### space_comments 表

```sql
CREATE TABLE space_comments (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  post_id     INTEGER NOT NULL,
  user_id     INTEGER NOT NULL,
  content     TEXT NOT NULL,
  created_at  DATETIME NOT NULL,
  FOREIGN KEY(post_id) REFERENCES space_posts(id),
  FOREIGN KEY(user_id) REFERENCES users(id)
);
```

### space_likes 表

```sql
CREATE TABLE space_likes (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  post_id     INTEGER NOT NULL,
  user_id     INTEGER NOT NULL,
  created_at  DATETIME NOT NULL,
  UNIQUE(post_id, user_id),
  FOREIGN KEY(post_id) REFERENCES space_posts(id),
  FOREIGN KEY(user_id) REFERENCES users(id)
);
```

---

## API 设计

### 认证

| Method | Path | 说明 |
|--------|------|------|
| POST | `/api/auth/register` | 注册 |
| POST | `/api/auth/login` | 登录（返回 JWT） |
| POST | `/api/auth/logout` | 登出 |
| GET | `/api/auth/me` | 当前用户信息 |

### 用户

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/users/search?q=` | 搜索用户 |
| GET | `/api/users/:id` | 查看用户资料 |
| PUT | `/api/users/me` | 更新个人信息 |
| PUT | `/api/users/me/avatar` | 上传头像 |

### 好友

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/friends` | 好友列表 |
| POST | `/api/friends/requests` | 发送好友申请 |
| GET | `/api/friends/requests` | 收到的申请列表 |
| PUT | `/api/friends/requests/:id` | 接受/拒绝申请 |
| DELETE | `/api/friends/:id` | 删除好友 |
| PUT | `/api/friends/:id/remark` | 修改备注 |
| PUT | `/api/friends/:id/group` | 修改分组 |

### 消息（私聊）

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/messages/:userId` | 获取与某用户聊天历史 |
| GET | `/api/conversations` | 会话列表（含最后一条消息） |
| PUT | `/api/messages/:userId/read` | 标记已读 |
| POST | `/api/messages/upload` | 上传图片/文件 |

### 群组

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/groups` | 我的群列表 |
| POST | `/api/groups` | 创建群 |
| GET | `/api/groups/:id` | 群详情 |
| GET | `/api/groups/:id/members` | 群成员列表 |
| POST | `/api/groups/:id/join` | 加入群 |
| DELETE | `/api/groups/:id/leave` | 退出群 |
| DELETE | `/api/groups/:id/members/:uid` | 踢人（仅群主） |
| DELETE | `/api/groups/:id` | 解散群（仅群主） |
| GET | `/api/groups/:id/messages` | 群聊历史 |
| GET | `/api/groups/search?q=` | 搜索群组 |

### QQ 空间

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/space/feed` | 动态列表（所有好友+自己） |
| GET | `/api/space/users/:id/posts` | 某用户的动态 |
| POST | `/api/space/posts` | 发布动态 |
| DELETE | `/api/space/posts/:id` | 删除动态 |
| POST | `/api/space/posts/:id/like` | 点赞 |
| DELETE | `/api/space/posts/:id/like` | 取消点赞 |
| POST | `/api/space/posts/:id/comments` | 评论 |
| DELETE | `/api/space/comments/:id` | 删除评论 |

### WebSocket

```
连接：WS /ws?token=<jwt>

客户端→服务端消息：
  { type: "chat_private", to_id: 123, content: {...}, msg_type: "text" }
  { type: "chat_group", group_id: 456, content: {...}, msg_type: "text" }
  { type: "heartbeat" }

服务端→客户端消息：
  { type: "message", data: {...} }           # 新消息
  { type: "friend_request", data: {...} }    # 好友申请
  { type: "friend_online", user_id: 123 }   # 好友上线
  { type: "friend_offline", user_id: 123 }  # 好友下线
  { type: "heartbeat_ack" }
```

---

## 开发阶段规划

### Phase 0：初始化 ✅（当前）
- [x] 创建项目目录结构
- [x] 编写 style-guide.md
- [x] 编写 dev-progress.md（本文件）
- [ ] 初始化前端 Vite + Vue3 项目
- [ ] 初始化后端 Go 模块

### Phase 1：基础骨架
- [ ] 后端：SQLite 连接 + 自动迁移所有表
- [ ] 后端：配置文件 (config.yaml) 结构
- [ ] 后端：JWT 认证中间件
- [ ] 后端：注册/登录/登出 API
- [ ] 前端：全局 CSS 变量应用
- [ ] 前端：三栏主布局（AppLayout + NavBar + ListPanel + ContentArea）
- [ ] 前端：登录/注册页面
- [ ] 前端：路由守卫（未登录跳转到 /login）

### Phase 2：好友系统
- [ ] 后端：好友 CRUD API（申请/同意/拒绝/删除）
- [ ] 前端：联系人列表（分组折叠/展开）
- [ ] 前端：搜索用户 + 发送好友申请弹窗
- [ ] 前端：好友申请通知（列表 + 实时 WS 推送）
- [ ] 前端：用户资料查看弹窗

### Phase 3：私聊
- [ ] 后端：WebSocket Hub 建立
- [ ] 后端：私聊消息存储 + 历史记录 API
- [ ] 前端：会话列表（含最新消息预览 + 未读角标）
- [ ] 前端：聊天窗口（气泡渲染）
- [ ] 前端：文字发送（回车/Ctrl+Enter）
- [ ] 前端：图片发送（上传 + 显示）
- [ ] 前端：消息历史滚动加载

### Phase 4：群组
- [ ] 后端：群组 CRUD API
- [ ] 后端：群消息存储 + WS 广播
- [ ] 前端：群列表
- [ ] 前端：群聊窗口（显示发送人名称）
- [ ] 前端：群成员侧边栏（群主可踢人）
- [ ] 前端：创建群 / 搜索加入群

### Phase 5：QQ 空间
- [ ] 后端：动态 CRUD + 点赞 + 评论 API
- [ ] 前端：空间 Feed 流（好友动态聚合）
- [ ] 前端：发布动态编辑器（文字 + 九宫格图片）
- [ ] 前端：个人主页（封面 + 动态列表）
- [ ] 前端：评论/点赞交互

### Phase 6：细节打磨
- [ ] 在线状态同步（WS 推送好友上下线）
- [ ] 未读消息统计（导航角标）
- [ ] 右键菜单（消息列表项、联系人）
- [ ] 时间格式化（今天/昨天/具体日期）
- [ ] 消息去重 + 乐观更新
- [ ] 加载状态 + 骨架屏
- [ ] 基础错误处理 + Toast 提示

---

## 配置文件说明

```yaml
# backend/config/config.yaml
server:
  port: 8080
  mode: debug        # debug / release

database:
  path: ./data/im.db # SQLite 文件路径

jwt:
  secret: "change-me-in-production"
  expire: 168h       # 7 天

upload:
  dir: ./data/uploads
  max_size: 10       # MB
  allowed_types:     # MIME 类型白名单
    - image/jpeg
    - image/png
    - image/gif
    - image/webp

cors:
  allow_origins:
    - "http://localhost:5173"
```

---

## 开发约定

### 前端约定

1. 所有 API 调用统一走 `src/api/` 层，禁止在 store/component 直接写 axios
2. 所有数据结构定义在 `src/types/`，禁止使用 `any`
3. 组件名使用 PascalCase，文件名同组件名
4. 组件拆分原则：超过 150 行或包含独立逻辑时拆出子组件
5. 颜色、间距必须使用 CSS 变量，禁止硬编码颜色值
6. 禁止在模板中写复杂表达式，复杂逻辑用 computed

### 后端约定

1. 所有 HTTP 响应使用统一格式：`{ code, message, data }`
2. 错误返回合理的 HTTP 状态码
3. 所有数据库操作通过 dao 层，service 层不直接操作 db
4. WebSocket 消息统一走 Hub 广播，禁止直接写 ws.conn
5. 上传文件校验 MIME 类型和大小，路径禁止 path traversal
6. 数据库查询使用参数化查询（GORM 自动处理）
