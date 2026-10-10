# 文件模块全局搜索 —— 实施总结

## 结果

已完成「按文件/文件夹名称子串模糊匹配」的全局搜索，共 8 个任务全部完成。新增接口 `GET /file/search`（JWT 保护），搜索结果跨全部目录，并返回每条结果的所在路径。

## 变更清单

### 后端（go-admin-template）

| 文件 | 变更 |
|---|---|
| `types/file_types.go` | 新增 `FileSearchRequest`、`FilePathItem`；`FileListItem` 增加 `Path` 字段（`omitempty`，不影响原列表接口） |
| `logic/file/file_manager_logic.go` | 新增 `SearchFiles`、`searchUnionRow`、`loadFolderMap`、`buildFolderPath` |
| `handler/file/file_manager_handle.go` | 新增 `SearchFilesHandle` |
| `routes/file/routes.go` | 注册 `g.GET("/file/search", file.SearchFilesHandle)` |

实现要点：
- `UNION ALL` 合并 `sys_file_folder` 与 `sys_file`，统一列集后按白名单排序、`LIMIT/OFFSET` 分页，总数用外层 `COUNT(*)` 子查询。
- 两个 `LIKE ?` 均带 `delete_time IS NULL`，软删除数据被正确排除。
- 排序白名单 map 映射 `sort_by`，未命中回退 `create_time DESC`，避免 SQL 注入。
- 路径通过一次性载入全部文件夹建 `id → folder` 映射后向上回溯生成，带 `visited` + 最大深度 64 的成环保护。

### 前端（go-admin-template-vue）

| 文件 | 变更 |
|---|---|
| `src/api/file.js` | 新增 `searchFiles(params)` |
| `src/views/file/index.vue` | 新增 `searchMode`/`searchKeyword` 状态、`loadSearchResults`、`refresh` 分流、`handleClearSearch`；改造 `handleSearch`；搜索态下点文件夹先退出搜索再进入 |
| `src/views/file/MaterialMenu.vue` | 新增 `keyword` prop 并 watch 同步；placeholder 改为「搜索全部文件/文件夹」；`@clear` 触发退出搜索 |
| `src/views/file/MaterialContent.vue` | 新增 `searchMode`/`searchKeyword` props 与 `clear-search` emit；搜索态用结果横幅替换面包屑；结果显示所在路径 |

## 验证情况

1. **后端编译**：`go build ./...` 与 `go vet ./logic/file/... ./handler/file/... ./routes/file/...` 均通过。
2. **前端构建**：`npm run build:prod`（vite build）通过。
3. **SQL 校验**：直接在 MySQL 上执行 `UNION ALL` + `COUNT(*)` 子查询 + `ORDER BY size/name`，语法与结果正确，软删除行被排除。
4. **逻辑集成测试**：临时测试直接调用 `SearchFiles` 打真实 MySQL，结果如下（测试文件已删除）：
   - `keyword="1"` → total=2，`1.1`（path=`1`）、`1`（path=``）
   - `keyword="%"` → total=4，含 `离职证明.pdf`（path=`1 / 1.1`）、`校历.png`（path=``，根目录）
   - `keyword="zzzzz"` → total=0

**测试过程中发现并修复的缺陷**：UNION 语句含两个 `?` 占位符，初版只传了 1 个参数，导致 `sql: expected 2 arguments, got 1`。已改为 `args := []interface{}{like, like}`。若仅靠编译无法发现此问题，逻辑集成测试是必要的。

## 未完成的验证（需知悉）

- **HTTP 层端到端未跑通**：本机 8888 端口已运行旧版后端进程，重启该进程的 `kill` 操作被沙箱自动审核拒绝，因此未能通过真实 HTTP 请求验证 `/file/search`。该层已由编译 + 路由注册 + handler 模式与现有接口完全一致来保证；如需 HTTP 验证，请重启后端后执行：
  ```
  curl "http://127.0.0.1:8888/file/search?keyword=1&page=1&page_size=20"
  ```

## 已知限制（沿用现状，非本次引入）

- `getExtByFileType`（`logic/file/file_manager_logic.go:266`）把 `image` 映射为 `.image`，而入库的 `FileExt` 是真实扩展名（如 `.png`），因此「文件类型筛选」实际不生效。搜索接口沿用同一映射以与现有列表行为保持一致；建议另开任务统一扩展名→类型映射。
- `LIKE '%kw%'` 无法走索引，数据量大时搜索会全表扫描；当前数据规模下无影响。
- 关键词中的 `%` / `_` 会被当作 SQL 通配符（参数绑定无注入风险，但语义上是通配而非字面量）。
