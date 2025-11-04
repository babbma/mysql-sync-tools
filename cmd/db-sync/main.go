package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/db-sync-tools/internal/config"
	"github.com/db-sync-tools/internal/database"
	"github.com/db-sync-tools/internal/logger"
	"github.com/db-sync-tools/internal/metadata"
	"github.com/db-sync-tools/internal/sync"
)

var (
	configPath = flag.String("config", "config.yaml", "配置文件路径")
	version    = flag.Bool("version", false, "显示版本信息")
)

const (
	Version = "1.0.0"
	AppName = "MySQL Database Sync Tool"
)

func main() {
	flag.Parse()

	// 显示版本信息
	if *version {
		fmt.Printf("%s v%s\n", AppName, Version)
		os.Exit(0)
	}

	// 加载配置
	cfg, err := config.LoadOrDefault(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志系统
	if err := logger.Init(cfg.Log.Level, cfg.Log.File, cfg.Log.Console); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("=================================")
	logger.Info("%s v%s", AppName, Version)
	logger.Info("=================================")

	// 运行同步
	if err := run(cfg); err != nil {
		logger.Fatal("同步失败: %v", err)
	}

	logger.Info("程序正常退出")
}

func run(cfg *config.Config) error {
	// 连接源数据库
	logger.Info("连接源数据库: %s@%s",
		cfg.Source.Username,
		cfg.Source.URL,
	)
	sourceDB, err := database.Connect(&cfg.Source)
	if err != nil {
		return fmt.Errorf("连接源数据库失败: %w", err)
	}
	defer sourceDB.Close()
	logger.Info("源数据库连接成功")

	// 连接目标数据库
	logger.Info("连接目标数据库: %s@%s",
		cfg.Target.Username,
		cfg.Target.URL,
	)
	targetDB, err := database.Connect(&cfg.Target)
	if err != nil {
		return fmt.Errorf("连接目标数据库失败: %w", err)
	}
	defer targetDB.Close()
	logger.Info("目标数据库连接成功")

	// 创建元数据管理器
	metaManager := metadata.NewMetadataManager(
		sourceDB,
		cfg.Sync.ExcludeTables,
		cfg.Sync.IncludeTables,
	)

	// 发现需要同步的表
	tables, err := metaManager.DiscoverTables()
	if err != nil {
		return fmt.Errorf("发现表失败: %w", err)
	}

	if len(tables) == 0 {
		logger.Warn("没有需要同步的表")
		return nil
	}

	// 创建同步器
	syncer := sync.NewSyncer(
		sourceDB,
		targetDB,
		cfg.Sync.BatchSize,
		cfg.Sync.TruncateBeforeSync,
		cfg.Sync.Timeout,
	)

	// 设置信号处理（支持优雅退出）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Warn("收到信号: %v，正在优雅退出...", sig)
		cancel()
	}()

	// 执行同步
	stats, err := syncer.SyncAll(ctx, tables, metaManager)
	if err != nil {
		return fmt.Errorf("同步过程出错: %w", err)
	}

	// 如果有失败的表，返回错误
	if stats.FailedTables > 0 {
		return fmt.Errorf("有 %d 个表同步失败", stats.FailedTables)
	}

	return nil
}
