# 文件模块全局搜索（file-global-search）

## 1. 背景与现状

文件模块当前是一个「按目录浏览」的资源管理器：

- 后端 `GetFileList`（`go-admin-template/logic/file/file_manager_logic.go:161`）虽然同时返回文件夹与文件，但**强制按 `parent_folder_id` 收窄**（`:177-183`），`keyword` 只在该目录内做 `LIKE`（`:186-189`）。
- 前端搜索框在 `go-admin-template-vue/src/views/file/MaterialMenu.vue:40`，回车后 emit `search`，`index.vue:183` 的 `handleSearch` 仅把 keyword 塞回当前目录的列表请求。
- `GetFolderList`（`logic/file/file_manager_logic.go:26`）同样按 `parent_id` 收窄。

**结论**：现有搜索是「当前目录内搜索」，不是全局搜索。文件数据没有 user/dept 维度的隔离（`sys_file`、`sys_file_folder` 均无归属过滤），因此「全局」= 跨全部文件夹，无需额外权限过滤。

## 2. 需求

在文件模块提供一个全局搜索：按**文件/文件夹名称的子串模糊匹配**，跨全部目录检索，并给出每条结果所在的位置路径，支持分页、类型筛选与排序。

### 2.1 场景与处理逻辑

1. 用户在搜索框输入关键词回车 → 进入「搜索模式」，展示跨目录的匹配结果。
2. 结果分页展示，文件夹与文件混排（沿用现有网格 UI）。
3. 每条结果显示其**所在路径**（面包屑文本），解决「全局结果不知道在哪」的问题。
4. 点击文件夹结果 → 退出搜索模式并进入该文件夹；点击文件结果 → 沿用现有下载/预览行为。
5. 清空搜索 → 退出搜索模式，恢复当前目录的正常浏览。
6. 搜索模式下切换文件类型 / 排序 → 重新执行搜索。

### 2.2 设计决策（已定，无需再确认）

| 决策点 | 选择 | 理由 |
|---|---|---|
| 结果承载方式 | 复用现有 `FileListItem` + 网格 UI，仅新增 `path` 字段 | 改动最小，前后端渲染逻辑全部复用 |
| 现有搜索框语义 | 由「当前目录内搜索」升级为「全局搜索」 | 用户要的就是全局搜索；目录内筛选已可由进入目录浏览替代 |
| 匹配方式 | SQL `LIKE '%kw%'`（两侧通配） | 需求明确为「子串模糊匹配」；MySQL 默认排序规则大小写不敏感，符合预期 |
| 跨表分页 | SQL `UNION ALL` + `ORDER BY` + `LIMIT/OFFSET` | 文件夹与文件分属两张表，需统一排序与分页；避免全量载入内存 |
| 路径来源 | 一次性载入全部文件夹建 id→folder 映射，向上回溯祖先 | 文件夹数量可控，一次查询即可，避免 N+1 |
| 文件类型筛选 | 复用现有 `getExtByFileType` 映射 | 与现有列表行为保持一致（见 §5 已知问题） |

### 2.3 不做的事（范围边界）

- 不改动 `GetFileList` / `GetFolderList` 的现有行为。
- 不做文件名高亮、拼音/首字母匹配、全文检索、搜索历史。
- 不引入新的归属/权限模型。

## 3. 后端设计

### 3.1 影响文件

| 文件 | 类型 | 说明 |
|---|---|---|
| `go-admin-template/types/file_types.go` | 修改 | 新增 `FileSearchRequest`、`FilePathItem`；`FileListItem` 增加 `Path` 字段 |
| `go-admin-template/logic/file/file_manager_logic.go` | 修改 | 新增 `SearchFiles`，新增 `buildFolderPath` 辅助函数 |
| `go-admin-template/handler/file/file_manager_handle.go` | 修改 | 新增 `SearchFilesHandle` |
| `go-admin-template/routes/file/routes.go` | 修改 | 注册 `GET /file/search` |

### 3.2 类型定义（types/file_types.go）

```go
// FileSearchRequest 全局搜索请求
type FileSearchRequest struct {
	Keyword   string  `form:"keyword" binding:"required,max=255" label:"搜索关键词"`
	FileType  *string `form:"file_type" label:"文件类型"`
	SortBy    string  `form:"sort_by" label:"排序方式"`
	Page      int     `form:"page" label:"页码"`
	PageSize  int     `form:"page_size" label:"每页条数"`
}

// FilePathItem 路径节点
type FilePathItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
```

`FileListItem` 增加一个字段（其余不动，保证列表接口响应不变）：

```go
type FileListItem struct {
	// ... 现有字段保持不变 ...
	Path []FilePathItem `json:"path,omitempty"` // 所在目录路径（不含自身），仅搜索结果返回
}
```

### 3.3 Logic：SearchFiles

签名：

```go
func SearchFiles(ctx *svc.ServiceContext, req *types.FileSearchRequest) (*types.FileListResponse, error)
```

实现要点：

1. **默认分页**：`Page < 1 → 1`，`PageSize < 1 → 20`，并对 `PageSize` 设上限（如 100）防止过大分页。
2. **UNION ALL 查询**（统一列集，便于排序分页）：

```sql
SELECT 'folder' AS type, id AS ref_id, folder_name AS name, '' AS ext,
       0 AS size, parent_id AS parent_id, create_time
FROM sys_file_folder
WHERE delete_time IS NULL AND folder_name LIKE ?
UNION ALL
SELECT 'asset' AS type, id AS ref_id, file_name AS name, file_ext AS ext,
       file_size AS size, folder_id AS parent_id, create_time
FROM sys_file
WHERE delete_time IS NULL AND file_name LIKE ?
  [AND file_ext = ?]   -- 仅当 file_type 命中 image/video/audio/document 时追加
```

3. **排序白名单**（禁止拼接用户输入，防注入）：

```go
orderMap := map[types.SortBy]string{
	types.SortByNameAsc:  "name ASC",
	types.SortByNameDesc: "name DESC",
	types.SortByTimeAsc:  "create_time ASC",
	types.SortByTimeDesc: "create_time DESC",
	types.SortBySizeAsc:  "size ASC",
	types.SortBySizeDesc: "size DESC",
}
order := orderMap[types.SortBy(req.SortBy)]
if order == "" {
	order = "create_time DESC"
}
```

4. **总数**：`SELECT COUNT(*) FROM (<union>) t`。
5. **分页数据**：`... ORDER BY <order> LIMIT ? OFFSET ?`。
6. **路径补全**：载入全部文件夹（`id, folder_name, parent_id`）构建映射，对每条结果的 `parent_id` 向上回溯，生成 `[]FilePathItem`；用 `visited` 集合 + 最大深度（如 64）防止脏数据成环死循环。
7. **组装**：文件夹 → `Type:"folder"` + `FolderID/FolderName/ParentFolderID`；文件 → `Type:"asset"` + `AssetID/FileName/FileURL/FileSize/FileExt/ParentFolderID/CreatedAt`，两者都填 `Path`。
8. **错误处理**：DB 失败时 `ctx.Log.Errorf("%+v", errors.WithStack(err))` 并返回 `errors.New("搜索失败")`，与现有风格一致。

`LIKE` 参数统一为 `"%" + req.Keyword + "%"`。

### 3.4 Handler（handler/file/file_manager_handle.go）

```go
// SearchFilesHandle 全局搜索文件/文件夹
func SearchFilesHandle(c *gin.Context) {
	var req types.FileSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}
	resp, err := file.SearchFiles(svc.NewServiceContext(c), &req)
	response.HandleResponse(c, resp, err)
}
```

### 3.5 路由（routes/file/routes.go）

在文件管理分组内新增（沿用 `JwtMiddleware`）：

```go
g.GET("/file/search", file.SearchFilesHandle)
```

## 4. 前端设计

### 4.1 影响文件

| 文件 | 类型 | 说明 |
|---|---|---|
| `go-admin-template-vue/src/api/file.js` | 修改 | 新增 `searchFiles(params)` |
| `go-admin-template-vue/src/views/file/index.vue` | 修改 | 搜索模式状态与数据流 |
| `go-admin-template-vue/src/views/file/MaterialMenu.vue` | 修改 | 搜索框语义改为全局，支持外部清空 |
| `go-admin-template-vue/src/views/file/MaterialContent.vue` | 修改 | 搜索模式横幅 + 结果路径展示 + 清除搜索 |

### 4.2 API

```js
// 全局搜索文件/文件夹
export function searchFiles(params) {
    return request({
        url: '/file/search',
        method: 'get',
        params
    })
}
```

### 4.3 index.vue 状态与数据流

新增状态：

```js
const searchMode = ref(false)
const searchKeyword = ref('')
```

数据流：

```
MaterialMenu 输入回车
  → emit('search', kw)
  → index.handleSearch(kw)
      kw 为空 → 退出搜索模式，loadFileList()
      kw 非空 → searchKeyword=kw, searchMode=true, page=1, loadSearchResults()
  → searchFiles({ keyword, file_type, sort_by, page, page_size })
  → res.data = { total, list } → fileList / total
```

关键处理函数：

```js
const handleSearch = (kw) => {
    const k = (kw || '').trim()
    if (!k) { handleClearSearch(); return }
    searchKeyword.value = k
    searchMode.value = true
    page.value = 1
    loadSearchResults()
}

const loadSearchResults = async () => {
    loading.value = true
    try {
        const res = await searchFiles({
            keyword: searchKeyword.value,
            file_type: fileType.value || undefined,
            sort_by: sortBy.value,
            page: page.value,
            page_size: pageSize.value
        })
        if (res.code === 0) {
            fileList.value = res.data.list
            total.value = res.data.total
        }
    } catch (e) {
        console.error('搜索失败:', e)
    } finally {
        loading.value = false
    }
}

const handleClearSearch = () => {
    searchMode.value = false
    searchKeyword.value = ''
    page.value = 1
    loadFileList()
}
```

改造点：

- `handleTypeChange` / `handleSortChange`：`searchMode` 为真时改调 `loadSearchResults()`，否则维持 `loadFileList()`。
- `handleFolderClick`：若在搜索模式，先退出搜索模式（清空关键词），再进入目标文件夹并 `loadFileList()`。
- `handlePageChange`：按 `searchMode` 分流到 `loadSearchResults()` 或 `loadFileList()`。
- 创建/重命名/删除/上传成功后的刷新：搜索模式下调用 `loadSearchResults()`，否则 `loadFileList()`。
- 新增 `handleClearSearch` 绑定到 `MaterialContent` 的 `clear-search`，并把 `searchMode` / `searchKeyword` 传入 `MaterialContent`、把 `keyword` 传入 `MaterialMenu`。

### 4.4 MaterialMenu.vue

- 新增 prop `keyword`（String），`watch` 后同步到 `searchKeyword`，使父组件清空时输入框同步清空。
- 搜索框 `placeholder` 改为「搜索全部文件/文件夹」，回车继续 emit `search`。
- 可加一个清空按钮（`clearable` 已有，配合 `@clear` emit `search` 空串即可触发退出搜索模式）。

### 4.5 MaterialContent.vue

- 新增 props：`searchMode`（Boolean）、`searchKeyword`（String）。
- 新增 emit：`clear-search`。
- 工具栏：`searchMode` 为真时，用搜索结果横幅替换面包屑：

```html
<div v-if="searchMode" class="search-banner">
    <span>搜索 “{{ searchKeyword }}”，共 {{ total }} 个结果</span>
    <el-button link type="primary" @click="emit('clear-search')">清除搜索</el-button>
</div>
```

- 列表项：`searchMode` 为真时，在名称下方展示路径文本（由 `item.path` 拼接）：

```html
<div v-if="searchMode && item.path && item.path.length" class="file-path">
    {{ item.path.map(p => p.name).join(' / ') }}
</div>
```

- 分页逻辑复用现有 `el-pagination`，无需改动。

## 5. 边界条件与异常处理

| 场景 | 处理 |
|---|---|
| 关键词为空/全空格 | 视为清除搜索，退出搜索模式 |
| 关键词含 `%` / `_` | 作为普通字符参与 `LIKE`（GORM 参数绑定，无注入风险；`%`/`_` 会被当通配符，属可接受行为） |
| 无匹配结果 | `total=0`，前端沿用 `el-empty` 空状态 |
| `page` 超过总页数 | 返回空列表，前端空状态，不报错 |
| `page_size` 过大 | 后端封顶（如 100） |
| 文件夹父子关系成环（脏数据） | 路径回溯加 `visited` + 最大深度保护，避免死循环 |
| 文件类型为 `other` | 复用现有 `getExtByFileType`，返回空串 → 不追加过滤（与现有列表行为一致） |
| DB 查询失败 | 记录日志，返回 `搜索失败` |
| 关键词超长 | `binding:"max=255"` 校验拦截 |

**已知问题（不在本次范围，需知悉）**：`getExtByFileType`（`logic/file/file_manager_logic.go:266`）把 `image` 映射为 `.image`，而 `UploadFile` 写入的 `FileExt` 是真实扩展名（如 `.png`），因此现有「文件类型筛选」实际不生效。本次搜索接口沿用同一映射以保持行为一致；若要修，应另开任务统一扩展名→类型映射。

## 6. 预期结果

- 输入关键词回车，跨全部目录返回名称包含该子串的文件夹与文件，分页正确、总数正确。
- 每条结果可看到所在路径，文件夹结果可点击直达。
- 清空搜索回到目录浏览，原有目录内浏览/上传/重命名/删除功能不受影响。
- 类型筛选与排序在搜索结果上生效（类型筛选受 §5 已知问题限制）。
- 新增接口 `GET /file/search` 受 JWT 中间件保护。
