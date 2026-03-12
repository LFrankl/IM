# QQ 风格 UI 设计规范

> 本文件是开发过程中的 UI 参考标准，所有前端实现必须遵循此规范。
> 参考对象：QQ NT（New Technology）桌面版 9.x，2023-2024 年版。

---

## 一、整体布局

### 三栏结构

```
┌──────┬────────────────┬─────────────────────────────────┐
│  导  │                │                                 │
│  航  │   列表区域     │         主内容区域              │
│  栏  │   (好友/群)    │    (聊天窗口 / 空间动态)        │
│      │                │                                 │
│ 68px │    280px       │         flex: 1 (最小 400px)    │
└──────┴────────────────┴─────────────────────────────────┘
```

- **导航栏**（Nav Column）：宽 68px，固定，包含功能入口图标 + 用户头像
- **列表栏**（List Column）：宽 280px，可滚动，展示会话列表 / 联系人 / 群组
- **内容栏**（Content Column）：flex 自适应，最小宽度 400px

### 窗口尺寸
- 最小窗口：960px × 620px
- 推荐默认：1280px × 800px

---

## 二、颜色体系

### 主色（QQ Blue）

| 变量名 | Hex | 用途 |
|--------|-----|------|
| `--qq-blue-primary` | `#1677FF` | 主按钮、激活状态、链接 |
| `--qq-blue-hover` | `#4096FF` | 主色悬停 |
| `--qq-blue-pressed` | `#0958D9` | 主色按下 |
| `--qq-blue-light` | `#E6F4FF` | 主色浅底（选中行背景） |
| `--qq-blue-lighter` | `#BAE0FF` | 主色更浅 |

### 中性色（亮色模式）

| 变量名 | Hex | 用途 |
|--------|-----|------|
| `--bg-body` | `#F0F0F0` | 整体页面背景 |
| `--bg-nav` | `#2C2C2C` | 左侧导航栏背景（深色） |
| `--bg-nav-hover` | `#3A3A3A` | 导航项悬停 |
| `--bg-nav-active` | `#1677FF` | 导航项激活 |
| `--bg-list` | `#F5F5F5` | 列表栏背景 |
| `--bg-list-item-hover` | `#EAEAEA` | 列表项悬停 |
| `--bg-list-item-active` | `#D9E8FF` | 列表项激活/选中 |
| `--bg-chat` | `#EDEDED` | 聊天区域背景（默认壁纸色） |
| `--bg-surface` | `#FFFFFF` | 卡片、输入框等表面色 |
| `--bg-input` | `#FFFFFF` | 输入框背景 |
| `--bg-input-toolbar` | `#F9F9F9` | 聊天输入工具栏背景 |

### 文字色

| 变量名 | Hex | 用途 |
|--------|-----|------|
| `--text-primary` | `#1A1A1A` | 主要文字 |
| `--text-secondary` | `#666666` | 次要文字（时间、副标题） |
| `--text-tertiary` | `#999999` | 占位文字、更次要说明 |
| `--text-inverse` | `#FFFFFF` | 深色背景上的文字 |
| `--text-link` | `#1677FF` | 链接文字 |
| `--text-nav` | `#C8C8C8` | 导航栏图标文字（未激活） |

### 边框色

| 变量名 | Hex | 用途 |
|--------|-----|------|
| `--border-light` | `#E5E5E5` | 轻量分割线 |
| `--border-normal` | `#D9D9D9` | 普通边框 |
| `--border-input` | `#D0D0D0` | 输入框边框 |

### 功能色

| 变量名 | Hex | 用途 |
|--------|-----|------|
| `--color-success` | `#52C41A` | 在线（绿色） |
| `--color-warning` | `#FAAD14` | 忙碌/请勿打扰 |
| `--color-error` | `#FF4D4F` | 错误、删除 |
| `--color-offline` | `#BFBFBF` | 离线 |
| `--color-bubble-self` | `#95EC69` | 自己发送气泡（QQ 绿，仿微信QQ混合风） |
| `--color-bubble-self-alt` | `#C6E2FF` | 自己发送气泡（QQ 蓝，更贴近 QQ 风格） |
| `--color-bubble-other` | `#FFFFFF` | 对方发送气泡 |
| `--color-badge` | `#FF4D4F` | 未读消息角标 |

> **注意**：QQ NT 默认使用蓝色气泡（`#C6E2FF`）作为自己消息的背景色，对方消息为白色气泡。

### 暗色模式（Dark Mode，可选扩展）

> 暂不强制实现，但预留 CSS 变量结构，后期可切换。

```css
[data-theme="dark"] {
  --bg-body: #1A1A1A;
  --bg-nav: #191919;
  --bg-list: #222222;
  --bg-chat: #292929;
  --bg-surface: #2C2C2C;
  --text-primary: #E8E8E8;
  --text-secondary: #AAAAAA;
  --border-light: #333333;
}
```

---

## 三、排版规范

### 字体家族

```css
font-family:
  "Microsoft YaHei UI",    /* Windows 微软雅黑 UI */
  "Microsoft YaHei",       /* 微软雅黑 */
  "PingFang SC",           /* macOS 苹方 */
  "Noto Sans CJK SC",      /* Linux */
  -apple-system,
  BlinkMacSystemFont,
  "Segoe UI",
  sans-serif;
```

### 字体规模

| 场景 | 大小 | 字重 | 行高 |
|------|------|------|------|
| 消息文字（主要） | 14px | 400 | 1.6 |
| 聊天对象名称（列表） | 14px | 500 | 1.4 |
| 消息预览（列表摘要） | 12px | 400 | 1.4 |
| 时间戳 | 11px | 400 | 1.2 |
| 分组标题 | 12px | 600 | 1.4 |
| 导航图标文字 | 10px | 400 | 1.2 |
| 输入框 | 14px | 400 | 1.6 |
| 标题栏联系人名 | 16px | 600 | 1.3 |
| 空间动态正文 | 14px | 400 | 1.8 |
| 按钮 | 13px | 500 | 1 |

---

## 四、间距系统

采用 4px 基础单位：

```
--spacing-xs:   4px
--spacing-sm:   8px
--spacing-md:  12px
--spacing-lg:  16px
--spacing-xl:  24px
--spacing-2xl: 32px
```

---

## 五、圆角规范

| 场景 | 值 |
|------|----|
| 头像（联系人/群组） | `8px`（圆角方形，QQ NT 风格） |
| 头像（用户本人小图） | `50%`（可配置） |
| 聊天气泡（通用） | `8px` |
| 气泡尖角侧 | `4px`（靠近头像侧较小） |
| 输入框 | `6px` |
| 按钮（普通） | `6px` |
| 按钮（小） | `4px` |
| 角标 | `10px`（圆形/胶囊） |
| 卡片 | `8px` |
| 弹窗 | `12px` |

---

## 六、核心组件规范

### 6.1 导航栏（NavBar）

```
宽度：68px，高度：100vh
背景：#2C2C2C（深色，与列表栏区分）

顶部：用户头像（40×40px，圆角8px，点击进入个人资料）
中部：功能入口图标（垂直排列）
  - 消息（默认激活）
  - 联系人
  - 群组
  - QQ 空间
底部：设置图标

图标规格：24×24px SVG
图标 + 文字标签：图标在上，文字（10px）在下，总高度约 48px
激活：图标背景变 #1677FF，文字变 #FFFFFF
未激活：图标颜色 #C8C8C8
悬停：背景 rgba(255,255,255,0.1)，圆角 8px
```

### 6.2 列表栏（ListPanel）

```
宽度：280px，高度：100vh
背景：#F5F5F5
顶部搜索栏高度：44px，背景 #EEEEEE

列表项高度：64px（含头像）
  - 头像：40×40px，左侧 margin 12px
  - 主文字（名称）：14px/500，单行溢出省略
  - 次文字（消息预览）：12px/400，color #999，单行溢出省略
  - 时间戳：11px，右上角，color #999
  - 未读角标：右下角，红色圆圈或胶囊

分组（联系人/群组列表时）：
  - 分组标题高度：28px，12px/600，#666
  - 展开/折叠箭头：12px，右侧

选中列表项：背景 #D9E8FF，左边 3px 蓝色条
悬停列表项：背景 #EAEAEA
```

### 6.3 聊天区域（ChatArea）

```
背景：#EDEDED（可设壁纸）

标题栏（ChatHeader）：
  高度：52px
  背景：#FFFFFF，底部 border 1px #E5E5E5
  内容：联系人名（16px/600）+ 在线状态文字 + 右侧操作图标

消息区（MessageList）：
  padding：16px 24px
  flex-direction：column（从上到下）

消息气泡（Bubble）：
  自己（靠右）：
    头像在右侧，bubble 在头像左
    背景：#C6E2FF（蓝）
    text：#1A1A1A
    max-width：60%
    border-radius：8px 2px 8px 8px（右上角小）

  对方（靠左）：
    头像在左侧，bubble 在头像右
    背景：#FFFFFF，box-shadow 0 1px 3px rgba(0,0,0,0.1)
    text：#1A1A1A
    max-width：60%
    border-radius：2px 8px 8px 8px（左上角小）

  气泡内 padding：10px 14px
  气泡间距（margin-bottom）：12px

  消息发送者名称：12px，#666，气泡上方（仅群聊显示）

时间分割线：
  居中，"上午 10:30" 格式，12px，#999
  间隔：相邻消息超过 5 分钟才显示

输入区（InputArea）：
  高度：min 120px，max 240px
  背景：#FFFFFF，顶部 border 1px #E5E5E5

  工具栏（toolbar）：高度 36px，背景 #F9F9F9
    工具图标：emoji、图片、文件、截图等，20×20px，#666

  文字输入框：14px，resize 竖向，padding 10px 16px

  发送按钮：右下角，蓝色圆角按钮，"发送" + Enter快捷键提示
```

### 6.4 头像（Avatar）

```
尺寸规格：
  - 大（个人主页）：80×80px
  - 中（聊天标题栏）：36×36px
  - 标准（列表）：40×40px
  - 小（消息气泡旁）：32×32px
  - 迷你（导航栏）：40×40px

圆角：8px（统一圆角方形，QQ NT 风格）
边框：无（设计简洁）
默认占位色：渐变，基于用户 ID 生成（蓝色系）

在线状态角标：
  位置：头像右下角
  尺寸：10×10px，白色边框 2px
  颜色：在线 #52C41A，离开 #FAAD14，离线 #BFBFBF
```

### 6.5 状态指示器

```
在线（Online）：#52C41A（绿色实心圆）
忙碌（Busy）：#FF7300（橙色实心圆）
离线（Offline）：#BFBFBF（灰色空心圆）
隐身（Invisible）：#BFBFBF（灰色虚线圆）
```

### 6.6 搜索框

```
高度：32px
背景：#E8E8E8（列表栏内）
圆角：16px（胶囊形）
padding：0 12px
font-size：13px，placeholder color #999
搜索图标：左侧，14px，#999
focus：背景 #FFFFFF，border 1px #1677FF
```

### 6.7 未读角标（Badge）

```
单数字（1-99）：18×18px 红色圆形，字体 11px 白色
双数字（10-99）：胶囊形，22px 高，内边距 4px
99+：固定显示 "99+"，胶囊形
位置：列表项右上角，或导航图标右上角
颜色：#FF4D4F
```

### 6.8 右键菜单（ContextMenu）

```
背景：#FFFFFF
border：1px #E5E5E5
border-radius：8px
box-shadow：0 8px 24px rgba(0,0,0,0.15)
min-width：140px

菜单项高度：36px
padding：0 16px
font-size：13px，color #1A1A1A
分割线：1px #F0F0F0，margin 4px 0
危险操作（删除/踢人）：color #FF4D4F
悬停：背景 #F0F6FF，color #1677FF
```

---

## 七、QQ 空间模块规范

```
整体背景：#F4F4F8（淡蓝紫）
顶部个人信息区：
  - 封面图：100% 宽，200px 高，渐变覆盖
  - 头像（左下，大）：80×80px，border 3px white
  - 昵称：20px/700，白色

动态 Feed 区：
  最大宽度：680px，居中
  动态卡片：
    背景：#FFFFFF，圆角 12px，margin-bottom 16px
    padding：16px
    头像 + 名称 + 时间：顶部
    正文：14px，行高 1.8，color #1A1A1A
    图片网格：最多9宫格，gap 4px，单图最大 100% 宽
    操作栏（点赞/评论）：底部，border-top 1px #F0F0F0

  评论区：
    缩进 12px，背景 #F9F9F9，圆角 4px
    评论人名：蓝色 #1677FF
```

---

## 八、动效规范

| 场景 | 动效 | 时长 |
|------|------|------|
| 列表项悬停背景 | ease | 100ms |
| 按钮点击反馈 | scale(0.96) | 80ms |
| 面板切换（左→右滑入） | ease-out translateX | 200ms |
| 气泡出现（新消息） | ease-out + opacity 0→1 | 150ms |
| 气泡弹入（自己发送） | ease-out + scale 0.8→1 | 120ms |
| 弹窗/模态框出现 | ease-out + opacity + scale 0.95→1 | 180ms |
| 右键菜单出现 | ease-out + opacity + translateY -8→0 | 100ms |

---

## 九、图标规范

- 使用 SVG 图标（推荐 `@iconify/vue` + `mdi` 或自定义 SVG 集）
- 导航栏图标：24px，stroke 风格，stroke-width 1.5
- 工具栏图标：20px
- 操作图标（按钮内）：16px
- 状态图标：12px

---

## 十、响应式说明

本项目仅为桌面 Web（>= 960px），不做移动端适配。
窗口缩放时：
- 导航栏和列表栏宽度固定
- 内容栏 flex 自适应
- 最小宽度 960px 时不再缩小（overflow hidden）

---

## CSS 变量汇总

```css
:root {
  /* 主色 */
  --qq-blue-primary: #1677FF;
  --qq-blue-hover: #4096FF;
  --qq-blue-pressed: #0958D9;
  --qq-blue-light: #E6F4FF;
  --qq-blue-lighter: #BAE0FF;

  /* 背景 */
  --bg-body: #F0F0F0;
  --bg-nav: #2C2C2C;
  --bg-nav-hover: rgba(255,255,255,0.1);
  --bg-nav-active: #1677FF;
  --bg-list: #F5F5F5;
  --bg-list-item-hover: #EAEAEA;
  --bg-list-item-active: #D9E8FF;
  --bg-chat: #EDEDED;
  --bg-surface: #FFFFFF;
  --bg-input: #FFFFFF;
  --bg-input-toolbar: #F9F9F9;

  /* 文字 */
  --text-primary: #1A1A1A;
  --text-secondary: #666666;
  --text-tertiary: #999999;
  --text-inverse: #FFFFFF;
  --text-link: #1677FF;
  --text-nav: #C8C8C8;

  /* 边框 */
  --border-light: #E5E5E5;
  --border-normal: #D9D9D9;
  --border-input: #D0D0D0;

  /* 功能色 */
  --color-success: #52C41A;
  --color-warning: #FAAD14;
  --color-error: #FF4D4F;
  --color-offline: #BFBFBF;
  --color-bubble-self: #C6E2FF;
  --color-bubble-other: #FFFFFF;
  --color-badge: #FF4D4F;

  /* 间距 */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 12px;
  --spacing-lg: 16px;
  --spacing-xl: 24px;
  --spacing-2xl: 32px;

  /* 圆角 */
  --radius-avatar: 8px;
  --radius-bubble: 8px;
  --radius-input: 6px;
  --radius-btn: 6px;
  --radius-card: 8px;
  --radius-modal: 12px;
  --radius-search: 16px;

  /* 阴影 */
  --shadow-bubble-other: 0 1px 3px rgba(0,0,0,0.1);
  --shadow-menu: 0 8px 24px rgba(0,0,0,0.15);
  --shadow-modal: 0 16px 48px rgba(0,0,0,0.2);
  --shadow-card: 0 2px 8px rgba(0,0,0,0.06);

  /* 布局 */
  --nav-width: 68px;
  --list-width: 280px;
  --header-height: 52px;
  --input-area-min-height: 120px;

  /* 字体 */
  --font-family: "Microsoft YaHei UI", "Microsoft YaHei", "PingFang SC",
    "Noto Sans CJK SC", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  --font-size-xs: 10px;
  --font-size-sm: 11px;
  --font-size-caption: 12px;
  --font-size-body: 14px;
  --font-size-title: 16px;
  --font-size-large: 20px;

  /* 过渡 */
  --transition-fast: 100ms ease;
  --transition-normal: 200ms ease;
  --transition-slow: 300ms ease;
}
```
