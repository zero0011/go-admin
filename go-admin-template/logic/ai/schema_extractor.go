package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-admin-template/config"
	pkgredis "go-admin-template/pkg/redis"

	"gorm.io/gorm"
)

const schemaCacheKey = "ai:schema"

type columnInfo struct {
	TableName   string `gorm:"column:TABLE_NAME"`
	ColumnName  string `gorm:"column:COLUMN_NAME"`
	ColumnType  string `gorm:"column:COLUMN_TYPE"`
	ColumnComment string `gorm:"column:COLUMN_COMMENT"`
}

// sensitiveColumns lists column names to exclude from the schema.
var sensitiveColumns = map[string]bool{
	"password": true,
	"token":    true,
	"secret":   true,
	"salt":     true,
}

// ExtractSchema returns a human-readable schema string for the current database.
// The result is cached in Redis for 1 hour.
func ExtractSchema(ctx context.Context, db *gorm.DB) (string, error) {
	// Try cache first
	cached, err := pkgredis.Client.Get(ctx, schemaCacheKey).Result()
	if err == nil && cached != "" {
		return cached, nil
	}

	var columns []columnInfo
	err = db.WithContext(ctx).Raw(`
		SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE, COLUMN_COMMENT
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME, ORDINAL_POSITION
	`, config.MysqlConf.Db).Scan(&columns).Error
	if err != nil {
		return "", fmt.Errorf("schema extractor: query columns: %w", err)
	}

	// Group columns by table
	tableMap := make(map[string][]columnInfo)
	tableOrder := []string{}
	for _, col := range columns {
		if sensitiveColumns[strings.ToLower(col.ColumnName)] {
			continue
		}
		if _, exists := tableMap[col.TableName]; !exists {
			tableOrder = append(tableOrder, col.TableName)
		}
		tableMap[col.TableName] = append(tableMap[col.TableName], col)
	}

	// Format as CREATE TABLE-like schema text
	var sb strings.Builder
	for _, table := range tableOrder {
		sb.WriteString(fmt.Sprintf("表名: %s\n字段:\n", table))
		for _, col := range tableMap[table] {
			comment := col.ColumnComment
			if comment == "" {
				comment = col.ColumnName
			}
			sb.WriteString(fmt.Sprintf("  - %s (%s) -- %s\n", col.ColumnName, col.ColumnType, comment))
		}
		sb.WriteString("\n")
	}

	schema := sb.String()

	// Cache in Redis
	schemaBytes, _ := json.Marshal(schema)
	pkgredis.Client.Set(ctx, schemaCacheKey, string(schemaBytes), time.Hour)

	return schema, nil
}
