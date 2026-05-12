package ai

import (
	"fmt"
	"regexp"
	"strings"
)

// forbiddenKeywords are SQL keywords that must not appear in the generated query.
var forbiddenKeywords = []string{
	"INSERT", "UPDATE", "DELETE", "DROP", "TRUNCATE",
	"CREATE", "ALTER", "EXEC", "EXECUTE", "GRANT", "REVOKE",
}

// sensitiveFieldPatterns matches forbidden field names in SQL.
var sensitiveFieldPattern = regexp.MustCompile(`(?i)\b(password|token|secret|salt)\b`)

// limitPattern checks whether the SQL already contains a LIMIT clause.
var limitPattern = regexp.MustCompile(`(?i)\bLIMIT\s+\d+`)

// ValidateAndClean validates the SQL for safety and injects LIMIT if missing.
// Returns the cleaned SQL or an error.
func ValidateAndClean(sql string, maxRows int) (string, error) {
	sql = strings.TrimSpace(sql)
	// Strip possible markdown code fences that the LLM might include
	sql = strings.TrimPrefix(sql, "```sql")
	sql = strings.TrimPrefix(sql, "```")
	sql = strings.TrimSuffix(sql, "```")
	sql = strings.TrimSpace(sql)

	upper := strings.ToUpper(sql)

	// Must start with SELECT
	fields := strings.Fields(upper)
	if len(fields) == 0 || fields[0] != "SELECT" {
		return "", fmt.Errorf("只允许 SELECT 查询，当前语句类型不合法")
	}

	// Check for forbidden keywords
	for _, kw := range forbiddenKeywords {
		// Use word-boundary style check: keyword surrounded by non-word chars
		pattern := regexp.MustCompile(`(?i)\b` + kw + `\b`)
		if pattern.MatchString(sql) {
			return "", fmt.Errorf("SQL 包含禁止的关键字: %s", kw)
		}
	}

	// Check for sensitive field access
	if sensitiveFieldPattern.MatchString(sql) {
		return "", fmt.Errorf("SQL 包含对敏感字段的查询，已被拒绝")
	}

	// Inject LIMIT if missing
	if !limitPattern.MatchString(sql) {
		// Remove trailing semicolons before appending LIMIT
		sql = strings.TrimRight(sql, "; ")
		sql = fmt.Sprintf("%s LIMIT %d", sql, maxRows)
	}

	return sql, nil
}
