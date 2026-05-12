package ai

import "fmt"

const promptTemplate = `你是一个 MySQL 专家。根据以下数据库表结构，将用户的自然语言问题转换为一条 SQL 查询语句。

【表结构】
%s

【约束规则】
1. 只生成 SELECT 语句，严禁 INSERT、UPDATE、DELETE、DROP、TRUNCATE、CREATE、ALTER 等操作
2. 不得查询 password、token、secret、salt 等敏感字段
3. 结果必须包含 LIMIT 子句，最大行数不超过 1000
4. 只输出 SQL 语句本身，不要任何解释、注释或 markdown 代码块标记

【问题】
%s`

// BuildPrompt constructs the LLM prompt with schema context.
func BuildPrompt(schema, question string) string {
	return fmt.Sprintf(promptTemplate, schema, question)
}
