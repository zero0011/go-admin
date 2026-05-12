package ai

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go-admin-template/config"
	"go-admin-template/model"
	"go-admin-template/pkg/llm"
	"go-admin-template/svc"
	"go-admin-template/types"
)

// Text2SQL converts a natural language question into SQL, executes it,
// and returns structured results with an inferred chart type.
func Text2SQL(ctx *svc.ServiceContext, req *types.Text2SQLRequest) (*types.Text2SQLResponse, error) {
	if len([]rune(req.Question)) > 500 {
		return nil, fmt.Errorf("问题长度不能超过 500 字符")
	}

	aiCfg := config.AIConf
	if aiCfg.ApiKey == "" || aiCfg.ApiKey == "your-api-key-here" {
		return nil, fmt.Errorf("AI 功能未配置，请在配置文件中设置 AI.ApiKey")
	}

	// 1. Extract schema (Redis cached)
	schema, err := ExtractSchema(context.Background(), model.DB())
	if err != nil {
		return nil, fmt.Errorf("获取数据库结构失败: %w", err)
	}

	// 2. Build prompt and call LLM
	prompt := BuildPrompt(schema, req.Question)
	client := &llm.Client{
		BaseURL:   aiCfg.BaseURL,
		APIKey:    aiCfg.ApiKey,
		Model:     aiCfg.Model,
		MaxTokens: aiCfg.MaxTokens,
	}

	maxRows := aiCfg.QueryMaxRows
	if maxRows <= 0 {
		maxRows = 1000
	}

	rawSQL, err := client.Complete(context.Background(), prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}

	// 3. Validate and clean SQL
	safeSQL, err := ValidateAndClean(rawSQL, maxRows)
	if err != nil {
		return nil, err
	}

	// 4. Execute SQL using raw query
	rows, err := model.DB().Raw(safeSQL).Rows()
	if err != nil {
		return nil, fmt.Errorf("SQL 执行失败: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列名失败: %w", err)
	}

	var resultRows [][]interface{}
	for rows.Next() {
		row := make([]interface{}, len(columns))
		rowPtrs := make([]interface{}, len(columns))
		for i := range row {
			rowPtrs[i] = &row[i]
		}
		if err = rows.Scan(rowPtrs...); err != nil {
			return nil, fmt.Errorf("读取行数据失败: %w", err)
		}
		// Convert []byte to string for JSON serialization
		converted := make([]interface{}, len(row))
		for i, v := range row {
			if b, ok := v.([]byte); ok {
				converted[i] = string(b)
			} else {
				converted[i] = v
			}
		}
		resultRows = append(resultRows, converted)
	}

	chartType := inferChartType(safeSQL, columns, resultRows)

	return &types.Text2SQLResponse{
		SQL:       safeSQL,
		Columns:   columns,
		Rows:      resultRows,
		ChartType: chartType,
	}, nil
}

// inferChartType determines the best visualization for the result.
func inferChartType(sql string, columns []string, rows [][]interface{}) string {
	colCount := len(columns)
	rowCount := len(rows)

	// Single value (e.g. COUNT(*))
	if colCount == 1 && rowCount == 1 {
		return "number"
	}

	// Two columns: label + value → pie or bar
	if colCount == 2 && rowCount > 0 {
		// Check if first column looks like a date/time for line chart
		firstCol := strings.ToLower(columns[0])
		if isTimeColumn(firstCol) || isTimeColumn(strings.ToLower(columns[1])) {
			return "line"
		}
		if rowCount <= 8 {
			return "pie"
		}
		return "bar"
	}

	// Check SQL for GROUP BY + time pattern → line
	upperSQL := strings.ToUpper(sql)
	if strings.Contains(upperSQL, "GROUP BY") {
		datePattern := regexp.MustCompile(`(?i)(DATE|YEAR|MONTH|DAY|WEEK|HOUR|created_at|updated_at|date_format)`)
		if datePattern.MatchString(sql) {
			return "line"
		}
	}

	return "table"
}

func isTimeColumn(name string) bool {
	timeWords := []string{"date", "time", "year", "month", "day", "week", "created", "updated", "at"}
	for _, w := range timeWords {
		if strings.Contains(name, w) {
			return true
		}
	}
	return false
}
