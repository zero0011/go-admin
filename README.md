# Go-Admin

通用简易后台管理系统，基于 **Gin + Vue3 + Element Plus**，适用于快速开发。

## 项目结构

```
go-admin/
├── go-admin-template/       # 后端 (Go/Gin)
└── go-admin-template-vue/   # 前端 (Vue3/Vite)
```

## 技术栈

### 后端
- [Gin](https://github.com/gin-gonic/gin) - HTTP 框架
- [Gorm](https://gorm.io) - ORM 框架
- [Casbin](https://github.com/casbin/casbin) - 权限认证
- [JWT](https://github.com/dgrijalva/jwt-go) - 登录认证

### 前端
- Vue 3
- Vite
- Element Plus
- Vue Router
- Pinia

## 功能特点

- 用户管理
- 角色管理
- 权限管理（RBAC）
- 登录认证（JWT）
- Swagger 接口文档

## 快速开始

### 后端

```bash
cd go-admin-template

# 配置数据库
# 编辑 etc/go-admin-template-local.yaml

# 导入数据库
mysql -u root -p < go-admin-template.sql

# 运行
go run go-admin-template.go -f etc/go-admin-template-local.yaml
```

### 前端

```bash
cd go-admin-template-vue

# 安装依赖
npm install

# 开发环境运行
npm run dev

# 生产构建
npm run build
```

## License

MIT


# Text-to-SQL 技术方案设计

## 一、功能目标

管理员在后台输入自然语言问题（如"上周新增了多少用户？"），系统自动转换为 SQL 并执行，以表格或图表形式返回结果。

---

## 二、整体架构

```
用户输入自然语言
        ↓
   [前端] 发送 POST /api/ai/text2sql
        ↓
   [后端] 构建 Prompt（注入表结构 + 业务说明）
        ↓
   [LLM API] 生成 SQL（OpenAI 兼容接口）
        ↓
   [后端] SQL 安全校验（仅允许 SELECT，过滤敏感字段）
        ↓
   [后端] 执行 SQL → 推断图表类型 → 返回结构化数据
        ↓
   [前端] 表格 / ECharts 图表动态渲染
```

---

## 三、后端设计（Go/Gin）

### 3.1 新增目录结构

```
go-admin-template/
├── pkg/
│   └── llm/
│       └── client.go              # LLM HTTP 客户端（OpenAI 兼容）
├── logic/
│   └── ai/
│       ├── schema_extractor.go    # 从 DB 提取表结构（Redis 缓存）
│       ├── prompt_builder.go      # 构建 LLM Prompt
│       ├── sql_validator.go       # SQL 安全校验
│       └── text2sql.go            # 业务逻辑编排
├── handler/
│   └── ai/
│       └── text2sql_handle.go     # 请求入口
└── routes/
    └── ai/
        └── routes.go              # 路由注册
```

### 3.2 API 接口定义

```
POST /api/ai/text2sql    （需 JWT 认证）

Request:
{
  "question": "上周新增了多少用户？"
}

Response:
{
  "sql":       "SELECT COUNT(*) AS cnt FROM admin_user WHERE ...",
  "columns":   ["cnt"],
  "rows":      [[42]],
  "chartType": "number"    // number / bar / line / pie / table
}
```

### 3.3 Prompt 构建策略

```
你是一个 MySQL 专家。根据以下表结构将自然语言问题转换为 SQL。

【表结构】
{schema}          ← 从 information_schema 动态提取，过滤 password/token 等敏感字段

【约束】
1. 只生成 SELECT 语句
2. 不得查询 password、token、secret、salt 字段
3. 结果必须包含 LIMIT，最大 1000
4. 只输出 SQL，不要任何解释

【问题】{question}
```

### 3.4 SQL 安全校验（sql_validator.go）

- 解析首个 token，非 SELECT 直接拒绝
- 关键字黑名单：`INSERT UPDATE DELETE DROP TRUNCATE CREATE ALTER EXEC`
- 敏感字段黑名单：`password token secret salt`
- 若 LLM 未生成 LIMIT，强制注入 `LIMIT 1000`

### 3.5 Schema 提取（schema_extractor.go）

```sql
SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE, COLUMN_COMMENT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = ?
  AND COLUMN_NAME NOT IN ('password','token','secret','salt')
ORDER BY TABLE_NAME, ORDINAL_POSITION
```

结果序列化为 JSON 缓存至 Redis，Key = `ai:schema`，TTL = 1h。

### 3.6 LLM 客户端（pkg/llm/client.go）

使用 OpenAI Chat Completions API 格式，支持通过配置切换模型：

```go
type Client struct { BaseURL, APIKey, Model string; MaxTokens int }
func (c *Client) Complete(ctx context.Context, prompt string) (string, error)
```

### 3.7 图表类型推断（后端逻辑）

| 结果特征 | chartType |
|---|---|
| 1 列 + 1 行（单个聚合值）| `number` |
| 2 列 + 行数 ≤ 8 | `pie` |
| 2 列 + 行数 > 8 | `bar` |
| 含时间列（date/time/year）| `line` |
| 其他多列结果 | `table` |

### 3.8 配置项（yaml 新增）

```yaml
AI:
  ApiKey: "${AI_API_KEY}"
  BaseURL: "https://api.openai.com/v1"
  Model: "gpt-4o-mini"
  MaxTokens: 512
  QueryMaxRows: 1000
```

---

## 四、前端设计（Vue3）

### 4.1 新增目录结构

```
go-admin-template-vue/src/
├── api/
│   └── ai.js                  # 接口调用（独立 axios 实例，30s 超时）
└── views/
    └── ai/
        └── TextToSQL.vue      # 主页面（输入区 + SQL展示 + 结果渲染）
```

### 4.2 页面交互设计

页面分三区：

- **输入区**：文本框 + 快捷问题标签 + 发送按钮
- **SQL 展示区**：代码块展示生成的 SQL（只读，帮助用户理解和学习）
- **结果区**：根据 `chartType` 动态渲染 Table / 数字卡片 / ECharts 图表

### 4.3 图表渲染（ECharts，vue-echarts）

| chartType | 渲染方式 |
|---|---|
| `number` | 大数字卡片 |
| `table` | el-table 分页展示 |
| `bar` | ECharts 柱状图 |
| `line` | ECharts 折线图 |
| `pie` | ECharts 饼图 |

### 4.4 路由注册

```js
{
  path: '/ai',
  component: Layout,
  meta: { title: 'AI 查询', icon: 'ai' },
  children: [{
    path: 'text2sql',
    component: () => import('@/views/ai/TextToSQL.vue'),
    meta: { title: '智能数据查询' }
  }]
}
```

---

## 五、安全边界汇总

| 风险点 | 防护措施 |
|---|---|
| LLM 生成恶意 SQL | 首 token 校验 + 关键字黑名单 |
| 查询敏感字段 | Schema 提取时过滤 + Prompt 约束 |
| 大批量查询 | 强制 LIMIT 1000 |
| 权限越权 | JWT 中间件鉴权 |
| Prompt 注入 | 用户输入长度限制 ≤ 500 字符 |

---

## 六、开发里程碑

- [x] 技术方案设计
- [ ] Step 1：LLM Client + Schema 提取 + Prompt Builder
- [ ] Step 2：SQL 校验 + 执行 + 图表类型推断
- [ ] Step 3：Handler + Route 注册
- [ ] Step 4：前端页面（输入 + SQL 展示 + Table/数字渲染）
- [ ] Step 5：ECharts 图表渲染（bar/line/pie）

## 未来可扩展的方向

### 方向一：SQL 修正对话（提升准确率）

当 SQL 执行失败或结果为空时，把错误信息反馈给 LLM，让它自动修正：

    text
    插入
    复制
    Round 1: 用户问 → LLM 生成 SQL → 执行失败
    Round 2: 把原问题 + 错误信息 + 原SQL → 再次请求 LLM 修正

最多重试 2 次，超出则提示用户换一种问法。这把准确率从"一次猜中"提升到"多轮收敛"。

### 方向二：查询历史 + 收藏

持久化用户的查询记录（问题 + 生成的 SQL + 结果摘要），支持：

*   历史查询一键重跑
*   收藏常用查询，生成"我的看板"
*   团队共享收藏查询（相当于轻量级报表系统）

### 方向三：Schema 变更主动失效

监听 DDL 变更事件（MySQL binlog 或定时 diff），表结构变更时主动删除 Redis 缓存的 Key，而不是等 TTL 过期。

### 方向四：用户可切换图表类型

后端推断 `chartType` 作为默认值，前端提供切换按钮（柱/折/饼/表），用户可手动覆盖。低成本的体验提升。

### 方向五：数据行级权限

在 SQL 执行前，根据当前用户的角色，自动向 SQL 注入 WHERE 条件（如 `AND dept_id = ?`），实现行级数据隔离。这是从"能用"到"可上生产"的关键一步。

### 方向六：可观测性

*   记录每次 LLM 调用的耗时、token 消耗、成功/失败状态
*   监控 SQL 执行时间，超过阈值告警
*   统计哪些问题类型失败率高 → 优化 Prompt 或补充 Schema 注释

### 方向七：Schema 语义增强

当前 Schema 只注入表结构，可以进一步注入**业务术语词典**：

    text
    插入
    复制
    【业务术语说明】
    - 活跃用户：最近 7 天内有登录记录的用户
    - 新用户：注册时间在最近 30 天内的用户

这样 LLM 对业务概念的理解更准确，SQL 质量显著提升。