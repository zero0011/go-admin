# 文件模块全局搜索实施计划

- [x] Task 1: 后端新增搜索相关类型定义
    - 1.1: 在 `types/file_types.go` 新增 `FileSearchRequest`（keyword 必填 max=255、file_type、sort_by、page、page_size）
    - 1.2: 在 `types/file_types.go` 新增 `FilePathItem`（id、name）
    - 1.3: 为 `FileListItem` 增加 `Path []FilePathItem` 字段（`json:"path,omitempty"`），不影响现有列表接口响应

- [x] Task 2: 后端实现 SearchFiles Logic
    - 2.1: 在 `logic/file/file_manager_logic.go` 新增 `SearchFiles(ctx, req)` 函数骨架
    - 2.2: 处理分页默认值（page<1→1，page_size<1→20）并设 page_size 上限 100
    - 2.3: 构建 `UNION ALL` 查询：文件夹按 `folder_name LIKE`，文件按 `file_name LIKE`，均带 `delete_time IS NULL`；file_type 命中 image/video/audio/document 时追加 `file_ext = ?`
    - 2.4: 用排序白名单 map 将 sort_by 映射为 `ORDER BY` 片段，未命中回退 `create_time DESC`
    - 2.5: 查询总数（`SELECT COUNT(*) FROM (<union>) t`）与分页数据（`ORDER BY ... LIMIT ? OFFSET ?`）
    - 2.6: 新增 `buildFolderPath` 辅助函数：载入全部文件夹建 id→folder 映射，向上回溯生成路径，带 visited 与最大深度保护
    - 2.7: 组装 `FileListResponse`（folder/asset 两类 item 均填充 Path），并按现有风格记录日志、返回 `搜索失败` 错误

- [x] Task 3: 后端新增 Handler 与路由
    - 3.1: 在 `handler/file/file_manager_handle.go` 新增 `SearchFilesHandle`（ShouldBindQuery → SearchFiles → HandleResponse）
    - 3.2: 在 `routes/file/routes.go` 的 JWT 分组内注册 `GET /file/search`

- [x] Task 4: 前端新增搜索 API
    - 4.1: 在 `src/api/file.js` 新增 `searchFiles(params)`，GET `/file/search`

- [x] Task 5: 前端 index.vue 接入搜索模式
    - 5.1: 新增 `searchMode`、`searchKeyword` 状态
    - 5.2: 新增 `loadSearchResults`，调用 `searchFiles` 并回填 `fileList`/`total`
    - 5.3: 改造 `handleSearch`：空关键词走清除，非空进入搜索模式并查询
    - 5.4: 新增 `handleClearSearch`：退出搜索模式、清空关键词、恢复目录列表
    - 5.5: `handleTypeChange`/`handleSortChange`/`handlePageChange` 及增删改/上传成功后的刷新按 `searchMode` 分流
    - 5.6: `handleFolderClick` 在搜索模式下先退出搜索模式再进入目标文件夹
    - 5.7: 向 `MaterialMenu` 传 `keyword`，向 `MaterialContent` 传 `searchMode`/`searchKeyword` 并绑定 `clear-search`

- [x] Task 6: 前端 MaterialMenu.vue 适配全局搜索
    - 6.1: 新增 `keyword` prop 并 watch 同步到本地输入框，实现外部清空
    - 6.2: 搜索框 placeholder 改为「搜索全部文件/文件夹」，`clearable` 的 `@clear` 触发空关键词搜索以退出搜索模式

- [x] Task 7: 前端 MaterialContent.vue 展示搜索结果
    - 7.1: 新增 `searchMode`/`searchKeyword` props 与 `clear-search` emit
    - 7.2: 搜索模式下用结果横幅（含结果数与「清除搜索」）替换面包屑
    - 7.3: 搜索模式下在文件/文件夹名称下方展示 `item.path` 拼接的所在路径

- [x] Task 8: 编译与联调验证
    - 8.1: 后端 `go build ./...` 通过
    - 8.2: 前端 `npm run build`（或 dev 启动）无编译错误
    - 8.3: 联调验证：跨目录子串匹配、分页总数、路径展示、文件夹结果点击跳转、清除搜索恢复目录浏览、空关键词与无结果场景
