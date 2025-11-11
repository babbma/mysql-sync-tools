package metadata

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/db-sync-tools/internal/database"
	"github.com/db-sync-tools/internal/logger"
)

// TableMetadata 表元数据
type TableMetadata struct {
	Name       string
	RowCount   int64
	Columns    []database.ColumnInfo
	PrimaryKey []string
	CreateSQL  string
	IsView     bool
}

// MetadataManager 元数据管理器
type MetadataManager struct {
	sourceDB      *database.DB
	excludeTables []string
	includeTables []string
}

// NewMetadataManager 创建元数据管理器
func NewMetadataManager(sourceDB *database.DB, excludeTables, includeTables []string) *MetadataManager {
	return &MetadataManager{
		sourceDB:      sourceDB,
		excludeTables: excludeTables,
		includeTables: includeTables,
	}
}

// DiscoverTables 发现需要同步的表
func (m *MetadataManager) DiscoverTables() ([]string, error) {
	logger.Info("开始发现数据库表...")

	// 获取所有表
	allTables, err := m.sourceDB.GetTables()
	if err != nil {
		return nil, fmt.Errorf("获取表列表失败: %w", err)
	}

	logger.Debug("数据库中共有 %d 个表", len(allTables))

	// 过滤表
	var filteredTables []string
	for _, table := range allTables {
		if m.shouldSyncTable(table) {
			filteredTables = append(filteredTables, table)
		} else {
			logger.Debug("跳过表: %s", table)
		}
	}

	logger.Info("发现 %d 个需要同步的表", len(filteredTables))

	return filteredTables, nil
}

// GetTableMetadata 获取表的元数据
func (m *MetadataManager) GetTableMetadata(tableName string) (*TableMetadata, error) {
	metadata := &TableMetadata{
		Name: tableName,
	}

	// 判断是否为视图
	if tableType, err := m.sourceDB.GetTableType(tableName); err == nil && strings.EqualFold(tableType, "VIEW") {
		metadata.IsView = true
	}

	// 获取行数（视图不统计）
	if !metadata.IsView {
		rowCount, err := m.sourceDB.GetRowCount(tableName)
		if err != nil {
			return nil, fmt.Errorf("获取表行数失败: %w", err)
		}
		metadata.RowCount = rowCount
	}

	// 获取列信息
	columns, err := m.sourceDB.GetTableColumns(tableName)
	if err != nil {
		return nil, fmt.Errorf("获取表结构失败: %w", err)
	}
	metadata.Columns = columns

	// 获取主键（视图无主键）
	if !metadata.IsView {
		primaryKey, err := m.sourceDB.GetPrimaryKey(tableName)
		if err != nil {
			logger.Warn("获取表 %s 的主键失败: %v", tableName, err)
		}
		metadata.PrimaryKey = primaryKey
	}

	// 获取创建SQL
	if metadata.IsView {
		viewSQL, err := m.sourceDB.GetCreateViewSQL(tableName)
		if err != nil {
			return nil, fmt.Errorf("获取视图定义失败: %w", err)
		}
		metadata.CreateSQL = viewSQL
	} else {
		createSQL, err := m.sourceDB.GetCreateTableSQL(tableName)
		if err != nil {
			logger.Warn("获取表 %s 的建表SQL失败: %v，将使用列信息自动生成", tableName, err)
			// 生成简单的建表SQL
			createSQL = m.generateCreateTableSQL(tableName, columns)
		}
		metadata.CreateSQL = createSQL
	}

	return metadata, nil
}

// generateCreateTableSQL 根据列信息生成简单的建表SQL
func (m *MetadataManager) generateCreateTableSQL(tableName string, columns []database.ColumnInfo) string {
	var colDefs []string
	for _, col := range columns {
		def := fmt.Sprintf("`%s` %s", col.Field, col.Type)
		if col.Null == "NO" {
			def += " NOT NULL"
		}
		if col.Default != "" {
			def += fmt.Sprintf(" DEFAULT '%s'", col.Default)
		}
		if col.Extra != "" {
			def += " " + col.Extra
		}
		colDefs = append(colDefs, def)
	}

	return fmt.Sprintf("CREATE TABLE `%s` (\n  %s\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4",
		tableName,
		strings.Join(colDefs, ",\n  "))
}

// shouldSyncTable 判断表是否应该同步
func (m *MetadataManager) shouldSyncTable(tableName string) bool {
	// 如果指定了包含列表，只同步包含列表中的表
	if len(m.includeTables) > 0 {
		for _, pattern := range m.includeTables {
			if matchPattern(tableName, pattern) {
				// 还需要检查是否在排除列表中
				return !m.isExcluded(tableName)
			}
		}
		return false
	}

	// 检查是否在排除列表中
	return !m.isExcluded(tableName)
}

// isExcluded 判断表是否被排除
func (m *MetadataManager) isExcluded(tableName string) bool {
	for _, pattern := range m.excludeTables {
		if matchPattern(tableName, pattern) {
			return true
		}
	}
	return false
}

// matchPattern 匹配表名模式（支持通配符 * 和 ?）
func matchPattern(tableName, pattern string) bool {
	// 使用 filepath.Match 进行模式匹配
	matched, err := filepath.Match(pattern, tableName)
	if err != nil {
		// 如果模式无效，进行精确匹配
		return tableName == pattern
	}
	return matched
}

// GetColumnNames 获取列名列表
func (m *TableMetadata) GetColumnNames() []string {
	names := make([]string, len(m.Columns))
	for i, col := range m.Columns {
		names[i] = col.Field
	}
	return names
}

// GetColumnNamesQuoted 获取带引号的列名列表（用于SQL）
func (m *TableMetadata) GetColumnNamesQuoted() []string {
	names := make([]string, len(m.Columns))
	for i, col := range m.Columns {
		names[i] = fmt.Sprintf("`%s`", col.Field)
	}
	return names
}

// GetPlaceholders 获取占位符字符串（用于INSERT语句）
func (m *TableMetadata) GetPlaceholders(rowCount int) string {
	singleRow := "(" + strings.Repeat("?,", len(m.Columns)-1) + "?)"
	if rowCount == 1 {
		return singleRow
	}
	return strings.Repeat(singleRow+",", rowCount-1) + singleRow
}

// HasPrimaryKey 是否有主键
func (m *TableMetadata) HasPrimaryKey() bool {
	return len(m.PrimaryKey) > 0
}
