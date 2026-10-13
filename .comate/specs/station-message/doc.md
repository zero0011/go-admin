# 站内信与实时公告（station-message）

> 本文档为需求文档（含关键设计决策），暂不进入编码。v1 范围已全部确认，见 §0。

## 0. 已确认决策（v1 冻结）

| 决策点 | 结论 |
|---|---|
| 推送通道 | **SSE** |
| 消息范围 | **仅运营广播公告**（不做系统事件触发、不做点对点私信） |
| 发布权限 | **复用现有角色，不新增权限点**（现状等价「仅超管可发布」） |
| 部署形态 | 未定 → **一次到位**：直接走 Redis Pub/Sub，单/多实例均可运行 |
| 撤回 | **不做**（v1 发布后不可撤回、不可编辑） |
| 内容形态 | **纯文本** |
| 提醒强度 | **不做分级、不做强提醒**（无弹窗/声音/桌面通知，仅徽标更新） |
| 目标人群 | **仅「全体」广播**（不做按角色/指定用户） |
| 放开给非超管 | **不需要**（维持仅超管可发布） |
| 历史保留 | 未特别要求 → v1 **不做自动清理**，长期保留 |

> ⚠️ **权限现状提示**：当前 SQL 种子中仅有一个角色「超级管理员」(id=1)，且 `etc/rbac_model.conf` 中 `role:1` 直接放行；`admin_casbin_rule` 无种子策略。因此「复用现有角色、不新增权限点」在**当前数据下等价于「仅超管可用」**——这已符合本次要求。

## 1. 背景与现状

### 1.1 项目技术栈
- 后端：Go 1.25 + Gin + GORM(MySQL) + Casbin(JWT/RBAC)，分层 `model → query → logic → handler → routes → types`，统一响应 `{code,data,message,requestId}`（`internal/response/response.go`）。
- 前端：Vue3 + Vite + Element Plus + Pinia + vue-router（hash），静态路由表 `src/router/index.js`。
- 缓存：Redis 已接入（`pkg/redis/db.go`，`github.com/go-redis/redis/v8`）。

### 1.2 现状盘点（与本需求相关的既有能力 / 空白）
| 事项 | 现状 | 对本需求的意义 |
|---|---|---|
| 通知 / 消息 / 公告模块 | **不存在**（`more.md` 中 `# 站内信，实时公告` 为待办） | 全新功能，无历史包袱 |
| WebSocket / SSE 实现 | **无**（`gin-contrib/sse` 仅作为间接依赖存在于 go.sum） | 需选定推送通道 |
| Redis | 已接入，支持 Pub/Sub | 可直接用于多实例推送扇出 |
| JWT 取 token | 支持 Header / Cookie / **Query 参数 `Authorization`**（`pkg/jwt/jwt.go:62`） | SSE 连接可直接复用现有 `JwtMiddleware` 鉴权 |
| 权限体系 | 后端 Casbin（`etc/rbac_model.conf`）+ `admin_auth` 菜单表；前端路由 `permission` 字段 + `user store` 按 `authkeys` 过滤 | 运营页面可零成本接入权限与菜单 |
| 顶部导航 | `Navbar.vue` 右侧 `.right-menu` | 通知铃铛的天然挂载点 |

**结论**：属于从 0 到 1 的新模块，但基础设施（Redis、JWT、RBAC、菜单、统一响应）齐备，无需引入新的框架级依赖。

## 2. 需求

### 2.1 功能需求（FR）
- **FR1 秒级实时通知**：运营发布公告后，在线用户在 **1 秒内**收到，铃铛未读数 +1。
- **FR2 运营推送页面**：超管通过平台页面撰写并发布公告（标题 + 纯文本正文），并可查看已发布列表。
- **FR3 一键已读**：用户可一键将全部未读消息标记为已读（幂等，返回本次标记数量）。
- **FR4 消息中心**：消息列表（区分已读/未读、分页）、未读数、单条已读、消息详情。
- **FR5 离线补投 / 断线补偿**：用户离线或连接中断期间产生的消息，在下次登录/重连时通过列表与未读数补齐，**保证不丢**。

### 2.2 非功能需求（NFR）
| 维度 | 要求 |
|---|---|
| 时效 | 发布→在线用户可见，P95 ≤ 1s |
| 可靠 | 消息以数据库为准，实时通道仅作加速；断线自动重连 + 重连后补偿拉取 |
| 并发 | 支持单实例数千长连接（goroutine-per-connection）；多实例下靠 Redis 扇出 |
| 安全 | 发布权限受控；正文纯文本，输出需转义防 XSS；连接鉴权且 token 过期即断开 |
| 兼容 | 依赖浏览器 `EventSource`；不支持时降级为轮询未读数 |
| 可观测 | 记录在线连接数、推送成功/失败数（可选，复用现有 Prometheus） |

### 2.3 范围边界（v1 不做）
- 不做系统事件触发消息、点对点私信会话、消息回复、附件收发。
- 不做撤回 / 编辑 / 草稿 / 定时发布。
- 不做消息分级、弹窗/声音/桌面通知等强提醒。
- 不做按角色 / 指定用户的定向推送（仅全体广播）。
- 不做已读统计报表；不改造现有权限模型（运营页复用现有后台管理权限，不新增权限点）。

## 3. 关键设计决策

### 3.1 推送通道：SSE
| 方案 | 优点 | 缺点 | 结论 |
|---|---|---|---|
| **SSE** | 单向 server→client 恰好匹配通知场景；`EventSource` 原生自动重连；Gin 原生支持；无需引入重依赖；`gin-contrib/sse` 已在依赖树 | 仅单向（回执走普通 REST 即可）；HTTP/1.1 同域 6 连接上限（单用户多标签页需注意）；需确认 Vite 代理不缓冲 | ✅ 采用 |
| WebSocket | 双向；无同域连接数限制 | 需引入 `gorilla/websocket`/`melody`；心跳/重连/鉴权全自管；项目无先例 | 备选 |
| 短轮询 | 实现最简单 | 无法兼顾「秒级」与低开销（要么延迟高要么请求多） | ❌ |

> 与 `more.md` 中「agent开发，**SSE推送**」的技术选型一致。

### 3.2 数据模型：**消息表 + 已读表**，广播不预写已读
- 全体广播**发布时不写入 N 条已读记录**（避免写放大）。
- 已读记录仅在用户**实际已读**时写入，`(message_id, user_id)` 唯一约束保证幂等。
- 因仅「全体」广播，**未读数 = 已发布消息总数 − 该用户已读记录数**（无需按目标人群过滤）。

### 3.3 一键已读实现
查出该用户所有未读消息 id → 批量 `INSERT ... ON DUPLICATE KEY UPDATE`（或 `INSERT IGNORE`）写已读记录 → 返回标记数量。幂等，重复点击不产生副作用。

### 3.4 多实例扇出：Redis Pub/Sub
发布流程：**写库 → `PUBLISH message:new` → 各实例订阅协程收到 → 查本机在线连接 → 推送**。
- 单实例本可用内存 Hub，但为免将来扩容返工，直接走 Redis（Redis 已在）。

### 3.5 权限接入（不新增权限点）
- **用户侧接口**（列表/未读数/已读/SSE）：仅挂 `JwtMiddleware`，所有登录用户可用。
- **运营侧接口**（发布/已发列表）：挂 `JwtMiddleware`，鉴权沿用现有角色，**不新增 `admin_auth` 权限点/菜单节点**。
- **前端**：运营页复用现有 `admin`（后台管理）权限可见性，挂在现有后台管理路由组下；`user store` 仍按 `authkeys` 过滤，无需改权限模型。
- 现状仅 `role:1` 超管可通行 → v1 即「仅超管可发布」，符合要求。

## 4. 数据模型（草案）

### 4.1 `sys_message`（公告）
| 字段 | 类型 | 说明 |
|---|---|---|
| id | int PK | |
| title | varchar(255) | 标题 |
| content | text | 正文（纯文本，输出需转义） |
| publisher_id | int | 发布人 |
| publish_time | datetime | 发布时间 |
| create_time / update_time / delete_time | datetime | 通用审计字段（软删除） |

### 4.2 `sys_message_read`（已读记录）
| 字段 | 类型 | 说明 |
|---|---|---|
| id | int PK | |
| message_id | int | 关联消息 |
| user_id | int | 用户 |
| read_time | datetime | 已读时间 |
| — | 唯一索引 | `(message_id, user_id)`，保证幂等 |

> 建表方式二选一（与项目现状一致）：加入 `model/mysql_setup.go` 的 `autoMigrate()`，或在 `go-admin-template.sql` 中补 DDL。现有 admin 表走 SQL dump，文件模块走 AutoMigrate。

## 5. 接口设计（草案）

### 5.1 用户侧（仅 JWT）
- `GET  /message/list` — 分页消息列表（含已读标记，可按已读/未读筛选）
- `GET  /message/unread-count` — 未读数（铃铛徽标）
- `POST /message/read` — 单条已读 `{id}`
- `POST /message/read-all` — **一键已读**，返回标记数量
- `GET  /message/stream` — **SSE 推送通道**（鉴权：`?Authorization=<jwt>`）

### 5.2 运营侧（JWT，复用现有角色）
- `POST /message/announcement` — 创建并发布公告
- `GET  /message/announcement/list` — 已发布列表

> 不提供编辑 / 撤回 / 删除接口。

### 5.3 SSE 事件约定（草案）
- 事件 `message`：新消息摘要（id/title/publish_time）
- 事件 `unread`：最新未读数
- 心跳：每 15–25s 发送一次，防代理超时断连
- 重连补偿：客户端连接成功（含重连）后主动拉一次 `unread-count` + `list` 首屏

## 6. 前端设计（草案）

| 位置 | 内容 |
|---|---|
| `src/components/MessageBell/index.vue` | 铃铛 + 未读徽标 + 下拉/抽屉消息列表 + 「全部已读」按钮 |
| `src/layout/components/Navbar.vue` | 在 `.right-menu` 挂载 MessageBell |
| `src/store/message.js` | Pinia：未读数、列表、SSE 连接状态 |
| `src/utils/sse.js` | EventSource 封装：登录后连接、指数退避重连、登出关闭、心跳监听 |
| `src/api/message.js` | 接口封装（对齐 `src/api/file.js` 风格） |
| `src/views/message/announcement/index.vue` | 运营发布页（标题 + 纯文本正文表单 + 已发列表） |
| `src/router/index.js` | 运营页挂到现有后台管理路由组下，复用 `permission: 'admin'`（不新增权限点） |

新消息到达：仅更新徽标未读数（不做弹窗/声音/桌面通知）。
SSE 地址：`/api/message/stream?Authorization=<token>`（Vite 代理 `/api`→8888 并去前缀，需确认 SSE 不被缓冲）。

## 7. 分期建议

- **P0（v1）**：全体广播公告（纯文本）+ SSE 实时 + 铃铛未读数 + 消息列表 + 单条已读 + **一键已读** + 运营发布页（复用后台管理权限）。
- **P1**：撤回 / 编辑 / 草稿 / 定时发布、已读统计、消息分级与强提醒、定向人群（角色/指定用户）、系统事件触发消息。
- **P2**：点对点私信、附件、多端。
