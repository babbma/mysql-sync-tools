package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
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
	useColor   bool
}

var defaultLogger *Logger

// Init 初始化日志系统
func Init(levelStr string, logFile string, console bool) error {
	level := parseLevel(levelStr)

	logger := &Logger{
		level:    level,
		useColor: console,
	}

	// 设置控制台日志
	if console {
		logger.consLogger = log.New(os.Stdout, "", 0)
	}

	// 设置文件日志
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("无法打开日志文件: %w", err)
		}
		logger.file = file
		logger.fileLogger = log.New(file, "", 0)
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

// log 记录日志
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
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
		l.fileLogger.Println(fileMsg)
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

