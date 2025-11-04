package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/db-sync-tools/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

// DB 数据库连接封装
type DB struct {
	*sql.DB
	Config *config.DatabaseConfig
}

// Connect 连接到数据库
func Connect(cfg *config.DatabaseConfig) (*DB, error) {
	dsn := cfg.GetDSN()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	return &DB{
		DB:     db,
		Config: cfg,
	}, nil
}

// GetTables 获取所有表名
func (db *DB) GetTables() ([]string, error) {
	query := "SHOW TABLES"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询表列表失败: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("扫描表名失败: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历表列表失败: %w", err)
	}

	return tables, nil
}

// GetTableColumns 获取表的列信息
func (db *DB) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf("SHOW COLUMNS FROM `%s`", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询表结构失败: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var null, key, extra sql.NullString
		var defaultVal sql.NullString

		if err := rows.Scan(&col.Field, &col.Type, &null, &key, &defaultVal, &extra); err != nil {
			return nil, fmt.Errorf("扫描列信息失败: %w", err)
		}

		col.Null = null.String
		col.Key = key.String
		col.Extra = extra.String
		if defaultVal.Valid {
			col.Default = defaultVal.String
		}

		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历列信息失败: %w", err)
	}

	return columns, nil
}

// GetRowCount 获取表的行数
func (db *DB) GetRowCount(tableName string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	var count int64
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询行数失败: %w", err)
	}
	return count, nil
}

// GetPrimaryKey 获取表的主键列
func (db *DB) GetPrimaryKey(tableName string) ([]string, error) {
	query := fmt.Sprintf("SHOW KEYS FROM `%s` WHERE Key_name = 'PRIMARY'", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询主键失败: %w", err)
	}
	defer rows.Close()

	// 获取列信息（动态适配不同MySQL版本）
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列信息失败: %w", err)
	}

	var primaryKeys []string
	for rows.Next() {
		// 创建足够的接收变量
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, fmt.Errorf("扫描主键信息失败: %w", err)
		}

		// Column_name 通常在第5列（索引4），但我们通过列名查找更安全
		for i, col := range columns {
			if col == "Column_name" && values[i] != nil {
				if colName, ok := values[i].([]byte); ok {
					primaryKeys = append(primaryKeys, string(colName))
				} else if colName, ok := values[i].(string); ok {
					primaryKeys = append(primaryKeys, colName)
				}
				break
			}
		}
	}

	return primaryKeys, nil
}

// TruncateTable 清空表
func (db *DB) TruncateTable(tableName string) error {
	query := fmt.Sprintf("TRUNCATE TABLE `%s`", tableName)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("清空表失败: %w", err)
	}
	return nil
}

// TableExists 检查表是否存在
func (db *DB) TableExists(tableName string) (bool, error) {
	query := fmt.Sprintf("SHOW TABLES LIKE '%s'", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return false, fmt.Errorf("检查表是否存在失败: %w", err)
	}
	defer rows.Close()

	return rows.Next(), nil
}

// CreateTableLike 根据源表创建相同结构的表
func (db *DB) CreateTableLike(tableName, sourceTableDDL string) error {
	// 先尝试删除表（如果存在）
	dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tableName)
	if _, err := db.Exec(dropQuery); err != nil {
		return fmt.Errorf("删除现有表失败: %w", err)
	}

	// 创建新表
	if _, err := db.Exec(sourceTableDDL); err != nil {
		return fmt.Errorf("创建表失败: %w", err)
	}

	return nil
}

// GetCreateTableSQL 获取建表SQL
func (db *DB) GetCreateTableSQL(tableName string) (string, error) {
	query := fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return "", fmt.Errorf("查询建表SQL失败: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", fmt.Errorf("未找到表 %s 的建表语句", tableName)
	}

	// 获取列信息（适配不同MySQL版本）
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("获取列信息失败: %w", err)
	}

	// 创建足够的接收变量
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	err = rows.Scan(valuePtrs...)
	if err != nil {
		return "", fmt.Errorf("扫描建表SQL失败: %w", err)
	}

	// 建表SQL通常在第2列（索引1）
	if len(values) < 2 {
		return "", fmt.Errorf("返回列数不足: %d", len(values))
	}

	var createSQL string
	if sqlBytes, ok := values[1].([]byte); ok {
		createSQL = string(sqlBytes)
	} else if sqlStr, ok := values[1].(string); ok {
		createSQL = sqlStr
	} else {
		return "", fmt.Errorf("无法解析建表SQL，类型: %T", values[1])
	}

	return createSQL, nil
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default string
	Extra   string
}
