# 站内信与实时公告（station-message）v1 实施计划

> 依据 `doc.md`（v1 冻结范围）：全体广播公告（纯文本）+ SSE 实时推送 + 铃铛未读数 + 消息列表 + 单条已读 + 一键已读 + 运营发布页（复用现有后台管理权限，不新增权限点）。
> 后端根目录 `go-admin-template/`，前端根目录 `go-admin-template-vue/`。

- [x] Task 1: 后端数据模型与自动迁移
    - 1.1: 新建 `model/message_models.go`，定义 `SysMessage`（表 `sys_message`：id/title/content/publisher_id/publish_time/create_time/update_time/delete_time）与 `SysMessageRead`（表 `sys_message_read`：id/message_id/user_id/read_time），均实现 `TableName()`，沿用 `gorm.DeletedAt` 软删除与 `SingularTable` 约定
    - 1.2: `SysMessageRead` 通过 `gorm:"uniqueIndex:uk_message_user"` 在 `message_id`+`user_id` 上建唯一索引，保证已读幂等
    - 1.3: 在 `model/mysql_setup.go` 的 `autoMigrate()` 中追加 `&SysMessage{}`、`&SysMessageRead{}`

- [x] Task 2: 后端 DTO 类型定义
    - 2.1: 新建 `types/message_types.go`，定义用户侧请求/响应：`MessageListRequest`（page/page_size/is_read 可选）、`MessageListItem`、`MessageListResponse`（total/list）、`MessageUnreadCountResponse`（count）、`MessageReadRequest`（id）、`MessageReadResponse`（count）
    - 2.2: 定义运营侧请求/响应：`MessageAnnouncementCreateRequest`（title 必填≤255、content 必填纯文本）、`MessageAnnouncementListItem`（id/title/publish_time/publisher_name）、`MessageAnnouncementListResponse`（total/list）
    - 2.3: 时间字段复用 `types.LocalTime` 以对齐现有 JSON 序列化

- [x] Task 3: 后端用户侧业务逻辑（logic/message）
    - 3.1: 新建 `logic/message/message_logic.go`，实现 `ListMessages`：分页查 `sys_message`（按 publish_time DESC），左连接/子查询 `sys_message_read` 计算每条 `is_read`，支持 is_read 筛选，返回 total/list
    - 3.2: 实现 `UnreadCount`：`已发布消息总数 − 该用户已读记录数`（因仅全体广播，无需按人群过滤）
    - 3.3: 实现 `MarkRead`：`INSERT ... ON DUPLICATE KEY UPDATE`/`INSERT IGNORE` 写单条已读记录，幂等
    - 3.4: 实现 `MarkAllRead`：查出该用户全部未读 message id → 批量写入已读记录 → 返回本次标记数量，重复调用返回 0 且无副作用
    - 3.5: 从 `svc.ServiceContext.GinContext` 取 `jwt.UserInfo` 解析当前用户 id（复用 `logic/file` 的 `getUserID` 写法）

- [x] Task 4: 后端运营侧业务逻辑（logic/message）
    - 4.1: 实现 `CreateAnnouncement`：写入 `sys_message`（title/content/publisher_id/publish_time=now）→ 事务提交后调用 SSE 发布函数 `PublishNewMessage(payload)`
    - 4.2: 实现 `AnnouncementList`：分页查询已发布公告（join `admin_user` 取发布人姓名），按 publish_time DESC
    - 4.3: 正文按纯文本处理，写库前不做富文本解析；输出转义交由前端渲染层负责

- [x] Task 5: 后端 SSE 推送基础设施（internal/sse 或 pkg/sse）
    - 5.1: 新建 SSE Hub：内存连接注册表（`map[userID]map[connID]chan []byte` + 读写锁），提供 `Register`/`Unregister`/`PushToUser`/`PushToAll` 方法
    - 5.2: 实现 Redis Pub/Sub 发布与订阅：发布端 `PUBLISH message:new <json>`；订阅协程 `SUBSCRIBE message:new`，收到后查本机在线连接并推送（多实例扇出）；单实例亦可用
    - 5.3: 实现 SSE 连接处理函数：设置 `Content-Type: text/event-stream`/`Cache-Control: no-cache`/`Connection: keep-alive`，使用 `c.SSEvent` + flush；监听 `c.Request.Context().Done()` 做断连清理
    - 5.4: 实现心跳：每 15–25s 发送一次 `: ping`/注释事件，防代理超时断连
    - 5.5: 定义事件约定：事件 `message`（新消息摘要 id/title/publish_time）、事件 `unread`（最新未读数）；连接建立后立即推送一次当前未读数

- [x] Task 6: 后端 Handler 层
    - 6.1: 新建 `handler/message/message_handle.go`，实现用户侧：`ListMessagesHandle`、`UnreadCountHandle`、`MarkReadHandle`、`MarkAllReadHandle`、`StreamHandle`（SSE）
    - 6.2: 实现运营侧：`CreateAnnouncementHandle`、`AnnouncementListHandle`
    - 6.3: 统一按 `ShouldBindQuery`/`ShouldBind`/`ShouldBindUri` → 调用 logic → `response.HandleResponse` 模式；SSE handler 不走统一响应（直接流式返回）

- [x] Task 7: 后端路由注册与启动装配
    - 7.1: 新建 `routes/message/routes.go`，`RegisterMessageRoute(e)`：用户侧仅挂 `middleware.JwtMiddleware`；运营侧挂 `JwtMiddleware + AuthMiddleware`（Casbin 沿用现有角色，role:1 旁路放行，不新增权限点）
    - 7.2: 注册路由：`GET /message/list`、`GET /message/unread-count`、`POST /message/read`、`POST /message/read-all`、`GET /message/stream`、`POST /message/announcement`、`GET /message/announcement/list`
    - 7.3: 在 `routes/setup.go` 的 `Setup()` 中调用 `messageRoutes.RegisterMessageRoute(e)`
    - 7.4: 在 `go-admin-template.go` 的 `main()` 中于 `model.Setup()` 之后启动 SSE Hub 与 Redis 订阅协程（`go sse.StartSubscriber()`）

- [x] Task 8: 前端 API 与 SSE 工具
    - 8.1: 新建 `src/api/message.js`，对齐 `src/api/file.js` 风格封装：`getMessageList`、`getUnreadCount`、`markRead`、`markAllRead`、`publishAnnouncement`、`getAnnouncementList`
    - 8.2: 新建 `src/utils/sse.js`：基于 `EventSource` 封装，URL 为 `/api/message/stream?Authorization=<token>`；实现登录后连接、指数退避重连、登出关闭、`message`/`unread` 事件回调、重连后补偿（主动拉一次 unread-count）

- [x] Task 9: 前端 Pinia store
    - 9.1: 新建 `src/store/message.js`（options 风格，对齐 `src/store/user.js`）：state 含 `unreadCount`/`list`/`connected`；actions 含 `fetchUnreadCount`、`fetchList`、`markRead`、`markAllRead`、`connect`/`disconnect`（内部调用 `utils/sse.js`）
    - 9.2: 新消息到达时仅更新 `unreadCount` 徽标（不做弹窗/声音/桌面通知）

- [x] Task 10: 前端铃铛组件与图标
    - 10.1: 在 `src/assets/svg/` 新增 `bell.svg` 图标（当前无该图标，路由/菜单依赖 svg-icons 自动注册）
    - 10.2: 新建 `src/components/MessageBell/index.vue`：铃铛 + `el-badge` 未读徽标 + 下拉/抽屉消息列表（区分已读/未读、分页）+ 单条已读 + 「全部已读」按钮
    - 10.3: 消息正文以纯文本渲染（`{{ }}` 插值，避免 `v-html`）防 XSS

- [x] Task 11: 前端 Navbar 挂载与生命周期联动
    - 11.1: 在 `src/layout/components/Navbar.vue` 的 `.right-menu` 中挂载 `<MessageBell />`
    - 11.2: 登录成功（`store/user.js` login/getUserInfo 后）建立 SSE 连接；登出时 `disconnect` 关闭连接

- [x] Task 12: 前端运营发布页与路由
    - 12.1: 新建 `src/views/message/announcement/index.vue`：标题 + 纯文本正文表单 + 发布按钮 + 已发布列表（分页）
    - 12.2: 在 `src/router/index.js` 现有「后台管理」`/admin` 路由组下新增子路由 `announcement`，复用 `permission: 'admin'`（不新增权限点）

- [x] Task 13: 编译与联调验证
    - 13.1: 后端 `go build ./...` 通过；确认 AutoMigrate 生成两张表
    - 13.2: 前端 `npm run build:prod` 通过（或 `npm run dev` 无报错）
    - 13.3: 手工验证：超管发布公告 → 在线用户 1s 内徽标 +1；消息列表/单条已读/一键已读幂等；断线重连后未读数与列表补偿不丢
