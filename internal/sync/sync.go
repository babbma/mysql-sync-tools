package sync

import (
	"context"
	"fmt"
	"strings"
	"time"

	syncstd "sync"

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
	maxConcurrency     int
}

// DecideObjectType 根据元数据判断对象类型（用于测试与分支决策）
func DecideObjectType(meta *metadata.TableMetadata) string {
	if meta != nil && meta.IsView {
		return "view"
	}
	return "table"
}

// NewSyncer 创建同步器
func NewSyncer(sourceDB, targetDB *database.DB, batchSize int, truncateBeforeSync bool, timeout int, concurrency int) *Syncer {
	return &Syncer{
		sourceDB:           sourceDB,
		targetDB:           targetDB,
		batchSize:          batchSize,
		truncateBeforeSync: truncateBeforeSync,
		timeout:            time.Duration(timeout) * time.Second,
		maxConcurrency:     concurrency,
	}
}

// SyncTable 同步单个表
func (s *Syncer) SyncTable(ctx context.Context, meta *metadata.TableMetadata) error {
	if DecideObjectType(meta) == "view" {
		logger.Info("开始同步视图: %s", meta.Name)
		// 视图：在目标端创建/替换视图定义，不做数据写入
		exists, err := s.targetDB.TableExists(meta.Name)
		if err != nil {
			return fmt.Errorf("检查目标视图是否存在失败: %w", err)
		}
		// 对于视图，先尝试删除
		if exists {
			logger.Info("目标存在对象 %s，尝试删除以重建视图", meta.Name)
			if _, err := s.targetDB.Exec(fmt.Sprintf("DROP VIEW IF EXISTS `%s`", meta.Name)); err != nil {
				return fmt.Errorf("删除目标视图失败: %w", err)
			}
		}
		// 元数据的 CreateSQL 为 SHOW CREATE VIEW 的结果，需要确保可直接执行
		if _, err := s.targetDB.Exec(meta.CreateSQL); err != nil {
			return fmt.Errorf("创建视图失败: %w", err)
		}
		logger.Info("视图 %s 同步完成", meta.Name)
		return nil
	}

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

	// 并发控制
	if s.maxConcurrency <= 0 {
		s.maxConcurrency = 1
	}
	sem := make(chan struct{}, s.maxConcurrency)
	var wg syncstd.WaitGroup
	var mu syncstd.Mutex

	for _, tableName := range tables {
		// 上下文取消检查
		select {
		case <-ctx.Done():
			logger.Warn("收集元数据过程被取消")
			return stats, fmt.Errorf("同步被取消")
		default:
		}

		meta, err := metaManager.GetTableMetadata(tableName)
		if err != nil {
			logger.Error("获取表 %s 的元数据失败: %v", tableName, err)
			mu.Lock()
			stats.FailedTables++
			stats.FailedDetails = append(stats.FailedDetails, FailedTableInfo{
				TableName: tableName,
				Error:     err.Error(),
			})
			mu.Unlock()
			continue
		}

		// 累计总行数
		mu.Lock()
		stats.TotalRows += meta.RowCount
		mu.Unlock()

		// 限流并发执行
		sem <- struct{}{}
		wg.Add(1)
		go func(m *metadata.TableMetadata) {
			defer wg.Done()
			defer func() { <-sem }()

			// 创建带超时的上下文
			tableCtx, cancel := context.WithTimeout(ctx, s.timeout)
			defer cancel()

			err := s.SyncTable(tableCtx, m)
			mu.Lock()
			if err != nil {
				logger.Error("同步对象 %s 失败: %v", m.Name, err)
				stats.FailedTables++
				stats.FailedDetails = append(stats.FailedDetails, FailedTableInfo{
					TableName: m.Name,
					Error:     err.Error(),
				})
			} else {
				stats.SyncedTables++
				stats.SyncedRows += m.RowCount
			}
			mu.Unlock()
		}(meta)
	}

	wg.Wait()

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
