package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Level 日志级别
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

var levelColors = map[Level]string{
	DEBUG: "\033[36m", // 青色
	INFO:  "\033[32m", // 绿色
	WARN:  "\033[33m", // 黄色
	ERROR: "\033[31m", // 红色
}

const colorReset = "\033[0m"

// Logger 日志记录器
type Logger struct {
	level      Level
	fileLogger *log.Logger
	consLogger *log.Logger
	file       *os.File
	logPath    string     // 主日志文件路径（固定名称）
	currentDay string     // 当前日志日期（YYYY-MM-DD）
	mu         sync.Mutex // 保护轮转与写入
	useColor   bool
}

var defaultLogger *Logger

// Init 初始化日志系统
func Init(levelStr string, logFile string, console bool) error {
	level := parseLevel(levelStr)

	logger := &Logger{
		level:    level,
		logPath:  logFile,
		useColor: console,
	}

	// 设置控制台日志
	if console {
		logger.consLogger = log.New(os.Stdout, "", 0)
	}

	// 设置文件日志
	if logFile != "" {
		// 确保目录存在
		if dir := filepath.Dir(logFile); dir != "" && dir != "." {
			_ = os.MkdirAll(dir, 0o755)
		}
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("无法打开日志文件: %w", err)
		}
		logger.file = file
		logger.fileLogger = log.New(file, "", 0)
		// 初始化当前日期
		logger.currentDay = time.Now().Format("2006-01-02")
	}

	defaultLogger = logger
	return nil
}

// Close 关闭日志系统
func Close() {
	if defaultLogger != nil && defaultLogger.file != nil {
		defaultLogger.file.Close()
	}
}

// parseLevel 解析日志级别
func parseLevel(levelStr string) Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// rotateIfNeeded 按日期轮转日志：当日期变化时将主日志重命名为带日期文件，并重新打开主日志
func (l *Logger) rotateIfNeeded(now time.Time) {
	if l.fileLogger == nil || l.logPath == "" {
		return
	}
	day := now.Format("2006-01-02")
	if day == l.currentDay {
		return
	}
	// 加锁保护轮转
	l.mu.Lock()
	defer l.mu.Unlock()
	// 双重检查
	if day == l.currentDay {
		return
	}

	// 关闭当前文件
	if l.file != nil {
		_ = l.file.Close()
	}

	// 生成备份文件名
	ext := filepath.Ext(l.logPath)
	base := strings.TrimSuffix(filepath.Base(l.logPath), ext)
	dir := filepath.Dir(l.logPath)
	if ext == "" {
		ext = ".log"
	}
	backup := filepath.Join(dir, fmt.Sprintf("%s.%s%s", base, l.currentDay, ext))

	// 如果备份文件已存在，附加时间避免覆盖
	if _, err := os.Stat(backup); err == nil {
		backup = filepath.Join(dir, fmt.Sprintf("%s.%s-%s%s", base, l.currentDay, now.Format("150405"), ext))
	}

	// 将主日志重命名为备份
	_ = os.Rename(l.logPath, backup)

	// 重新打开主日志
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}
	file, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// 发生异常时，回退到仅控制台输出
		l.file = nil
		l.fileLogger = nil
		l.currentDay = day
		return
	}
	l.file = file
	l.fileLogger = log.New(file, "", 0)
	l.currentDay = day
}

// log 记录日志
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	// 检查是否需要按日期轮转
	l.rotateIfNeeded(time.Now())

	levelName := levelNames[level]
	message := fmt.Sprintf(format, args...)

	// 控制台输出（带颜色）
	if l.consLogger != nil {
		var consoleMsg string
		if l.useColor {
			color := levelColors[level]
			consoleMsg = fmt.Sprintf("%s [%s%s%s] %s", timestamp, color, levelName, colorReset, message)
		} else {
			consoleMsg = fmt.Sprintf("%s [%s] %s", timestamp, levelName, message)
		}
		l.consLogger.Println(consoleMsg)
	}

	// 文件输出（不带颜色）
	if l.fileLogger != nil {
		fileMsg := fmt.Sprintf("%s [%s] %s", timestamp, levelName, message)
		// 文件写入加锁，避免与轮转并发
		l.mu.Lock()
		l.fileLogger.Println(fileMsg)
		l.mu.Unlock()
	}
}

// Debug 调试日志
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(DEBUG, format, args...)
	}
}

// Info 信息日志
func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(INFO, format, args...)
	}
}

// Warn 警告日志
func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(WARN, format, args...)
	}
}

// Error 错误日志
func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(ERROR, format, args...)
	}
}

// Fatal 致命错误日志（记录后退出程序）
func Fatal(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(ERROR, format, args...)
	}
	os.Exit(1)
}

// GetWriter 获取日志写入器（用于其他库的日志输出）
func GetWriter() io.Writer {
	if defaultLogger != nil && defaultLogger.file != nil {
		if defaultLogger.consLogger != nil {
			return io.MultiWriter(defaultLogger.file, os.Stdout)
		}
		return defaultLogger.file
	}
	return os.Stdout
}

// TestSetCurrentDay 测试辅助：设置当前日志日期（仅用于单元测试）
func TestSetCurrentDay(day string) {
	if defaultLogger != nil {
		defaultLogger.currentDay = day
	}
}
