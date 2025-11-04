package sync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/db-sync-tools/internal/database"
	"github.com/db-sync-tools/internal/logger"
	"github.com/db-sync-tools/internal/metadata"
)

// Syncer 数据同步器
type Syncer struct {
	sourceDB           *database.DB
	targetDB           *database.DB
	batchSize          int
	truncateBeforeSync bool
	timeout            time.Duration
}

// NewSyncer 创建同步器
func NewSyncer(sourceDB, targetDB *database.DB, batchSize int, truncateBeforeSync bool, timeout int) *Syncer {
	return &Syncer{
		sourceDB:           sourceDB,
		targetDB:           targetDB,
		batchSize:          batchSize,
		truncateBeforeSync: truncateBeforeSync,
		timeout:            time.Duration(timeout) * time.Second,
	}
}

// SyncTable 同步单个表
func (s *Syncer) SyncTable(ctx context.Context, meta *metadata.TableMetadata) error {
	logger.Info("开始同步表: %s (总行数: %d)", meta.Name, meta.RowCount)

	// 检查目标表是否存在
	exists, err := s.targetDB.TableExists(meta.Name)
	if err != nil {
		return fmt.Errorf("检查目标表是否存在失败: %w", err)
	}

	// 如果表不存在，创建表
	if !exists {
		logger.Info("目标表不存在，创建表: %s", meta.Name)
		if err := s.targetDB.CreateTableLike(meta.Name, meta.CreateSQL); err != nil {
			return fmt.Errorf("创建表失败: %w", err)
		}
	} else if s.truncateBeforeSync {
		// 如果需要，清空目标表
		logger.Info("清空目标表: %s", meta.Name)
		if err := s.targetDB.TruncateTable(meta.Name); err != nil {
			return fmt.Errorf("清空表失败: %w", err)
		}
	}

	// 如果表为空，跳过数据同步
	if meta.RowCount == 0 {
		logger.Info("表 %s 为空，跳过数据同步", meta.Name)
		return nil
	}

	// 开始同步数据
	startTime := time.Now()
	totalSynced := int64(0)

	// 计算总批次
	totalBatches := (meta.RowCount + int64(s.batchSize) - 1) / int64(s.batchSize)

	// 分批读取和写入数据
	for offset := int64(0); offset < meta.RowCount; offset += int64(s.batchSize) {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return fmt.Errorf("同步被取消")
		default:
		}

		currentBatch := offset/int64(s.batchSize) + 1

		// 读取一批数据
		rows, err := s.readBatch(meta.Name, meta.GetColumnNamesQuoted(), offset, s.batchSize)
		if err != nil {
			return fmt.Errorf("读取数据失败 (offset: %d): %w", offset, err)
		}

		if len(rows) == 0 {
			break
		}

		// 写入数据
		if err := s.writeBatch(meta, rows); err != nil {
			return fmt.Errorf("写入数据失败 (offset: %d): %w", offset, err)
		}

		totalSynced += int64(len(rows))

		// 计算进度和速度
		elapsed := time.Since(startTime).Seconds()
		speed := float64(totalSynced) / elapsed
		progress := float64(totalSynced) / float64(meta.RowCount) * 100

		logger.Info("表 %s: 批次 %d/%d, 进度: %.2f%% (%d/%d), 速度: %.0f 行/秒",
			meta.Name, currentBatch, totalBatches, progress, totalSynced, meta.RowCount, speed)
	}

	elapsed := time.Since(startTime)
	logger.Info("表 %s 同步完成: 共同步 %d 行, 耗时 %v", meta.Name, totalSynced, elapsed)

	return nil
}

// readBatch 读取一批数据
func (s *Syncer) readBatch(tableName string, columns []string, offset int64, limit int) ([][]interface{}, error) {
	query := fmt.Sprintf("SELECT %s FROM `%s` LIMIT %d OFFSET %d",
		strings.Join(columns, ", "),
		tableName,
		limit,
		offset,
	)

	rows, err := s.sourceDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %w", err)
	}
	defer rows.Close()

	// 获取列类型
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("获取列类型失败: %w", err)
	}

	var results [][]interface{}

	for rows.Next() {
		// 创建扫描目标
		values := make([]interface{}, len(columnTypes))
		valuePtrs := make([]interface{}, len(columnTypes))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描行
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("扫描行失败: %w", err)
		}

		// 处理特殊类型
		processedValues := make([]interface{}, len(values))
		for i, v := range values {
			processedValues[i] = processValue(v)
		}

		results = append(results, processedValues)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历行失败: %w", err)
	}

	return results, nil
}

// writeBatch 写入一批数据
func (s *Syncer) writeBatch(meta *metadata.TableMetadata, rows [][]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	// 构建INSERT语句
	columns := meta.GetColumnNamesQuoted()
	columnsPart := strings.Join(columns, ", ")

	// 构建占位符
	singlePlaceholder := "(" + strings.Repeat("?,", len(columns)-1) + "?)"
	placeholders := make([]string, len(rows))
	for i := range placeholders {
		placeholders[i] = singlePlaceholder
	}
	placeholdersPart := strings.Join(placeholders, ",")

	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s",
		meta.Name,
		columnsPart,
		placeholdersPart,
	)

	// 展开所有值
	var values []interface{}
	for _, row := range rows {
		values = append(values, row...)
	}

	// 执行插入
	_, err := s.targetDB.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("执行INSERT失败: %w", err)
	}

	return nil
}

// processValue 处理特殊类型的值
func processValue(value interface{}) interface{} {
	// 处理字节数组
	if b, ok := value.([]byte); ok {
		return string(b)
	}

	// 处理NULL值
	if value == nil {
		return nil
	}

	return value
}

// SyncStats 同步统计信息
type SyncStats struct {
	TotalTables   int
	SyncedTables  int
	FailedTables  int
	TotalRows     int64
	SyncedRows    int64
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	FailedDetails []FailedTableInfo
}

// FailedTableInfo 失败表信息
type FailedTableInfo struct {
	TableName string
	Error     string
}

// SyncAll 同步所有表
func (s *Syncer) SyncAll(ctx context.Context, tables []string, metaManager *metadata.MetadataManager) (*SyncStats, error) {
	stats := &SyncStats{
		TotalTables:   len(tables),
		StartTime:     time.Now(),
		FailedDetails: []FailedTableInfo{},
	}

	logger.Info("==============================")
	logger.Info("开始同步数据库")
	logger.Info("源数据库: %s@%s",
		s.sourceDB.Config.Username,
		s.sourceDB.Config.URL,
	)
	logger.Info("目标数据库: %s@%s",
		s.targetDB.Config.Username,
		s.targetDB.Config.URL,
	)
	logger.Info("批量大小: %d", s.batchSize)
	logger.Info("总表数: %d", len(tables))
	logger.Info("==============================")

	// 首先收集所有表的元数据和总行数
	var tableMetas []*metadata.TableMetadata
	for _, tableName := range tables {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			logger.Warn("收集元数据过程被取消")
			return stats, fmt.Errorf("同步被取消")
		default:
		}

		meta, err := metaManager.GetTableMetadata(tableName)
		if err != nil {
			logger.Error("获取表 %s 的元数据失败: %v", tableName, err)
			stats.FailedTables++
			stats.FailedDetails = append(stats.FailedDetails, FailedTableInfo{
				TableName: tableName,
				Error:     err.Error(),
			})
			continue
		} else {
			logger.Info("获取表 %s 的元数据成功", tableName)
			logger.Info("表名: %s", meta.Name)
			logger.Info("行数: %d", meta.RowCount)
			logger.Info("列数: %d", len(meta.Columns))
			logger.Info("主键: %v", meta.PrimaryKey)
			logger.Info("建表SQL: %s", meta.CreateSQL)
		}
		tableMetas = append(tableMetas, meta)
		stats.TotalRows += meta.RowCount
	}

	logger.Info("总数据行数: %d", stats.TotalRows)

	// 逐个同步表
	for i, meta := range tableMetas {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			logger.Warn("同步过程被取消")
			stats.EndTime = time.Now()
			stats.Duration = stats.EndTime.Sub(stats.StartTime)
			return stats, nil
		default:
		}

		logger.Info("正在同步第 %d/%d 个表...", i+1, len(tableMetas))

		// 创建带超时的上下文
		tableCtx, cancel := context.WithTimeout(ctx, s.timeout)

		err := s.SyncTable(tableCtx, meta)
		cancel() // 及时释放资源

		if err != nil {
			logger.Error("同步表 %s 失败: %v", meta.Name, err)
			stats.FailedTables++
			stats.FailedDetails = append(stats.FailedDetails, FailedTableInfo{
				TableName: meta.Name,
				Error:     err.Error(),
			})
		} else {
			stats.SyncedTables++
			stats.SyncedRows += meta.RowCount
		}
	}

	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime)

	// 打印统计信息
	logger.Info("==============================")
	logger.Info("数据库同步完成!")
	logger.Info("总表数: %d", stats.TotalTables)
	logger.Info("成功: %d", stats.SyncedTables)
	logger.Info("失败: %d", stats.FailedTables)
	logger.Info("总行数: %d", stats.TotalRows)
	logger.Info("同步行数: %d", stats.SyncedRows)
	logger.Info("总耗时: %v", stats.Duration)
	if stats.Duration.Seconds() > 0 {
		logger.Info("平均速度: %.0f 行/秒", float64(stats.SyncedRows)/stats.Duration.Seconds())
	}
	logger.Info("==============================")

	if stats.FailedTables > 0 {
		logger.Warn("以下表同步失败:")
		for _, failed := range stats.FailedDetails {
			logger.Warn("  - %s: %s", failed.TableName, failed.Error)
		}
	}

	return stats, nil
}
