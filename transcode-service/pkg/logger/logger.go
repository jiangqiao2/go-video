package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
	"transcode-service/pkg/assert"
	"transcode-service/pkg/config"
)

var (
	loggerOnce      sync.Once
	singletonLogger *Logger
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String 返回日志级别字符串
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry 日志条目结构
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
	Function  string                 `json:"function,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
}

// Logger 日志服务
type Logger struct {
	level      LogLevel
	format     string // "json" or "text"
	output     string // "stdout", "stderr", or file path
	file       *os.File
	mutex      sync.RWMutex
	enableFile bool
}

// DefaultLogger 获取默认日志服务单例
func DefaultLogger() *Logger {
	assert.NotCircular()
	loggerOnce.Do(func() {
		// 使用默认配置
		singletonLogger = &Logger{
			level:      INFO,
			format:     "json",
			output:     "stdout",
			enableFile: false,
		}
	})
	assert.NotNil(singletonLogger)
	return singletonLogger
}

// NewLogger 创建新的日志服务实例（支持依赖注入）
func NewLogger(cfg *config.Config) *Logger {
	logger := &Logger{
		level:      parseLogLevel(cfg.Log.Level),
		format:     cfg.Log.Format,
		output:     cfg.Log.Output,
		enableFile: cfg.Log.Output == "file",
	}

	// 如果配置为文件输出，打开文件
	if logger.enableFile && cfg.Log.Filename != "" {
		file, err := os.OpenFile(cfg.Log.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("无法打开日志文件 %s: %v，使用标准输出", cfg.Log.Filename, err)
			logger.output = "stdout"
			logger.enableFile = false
		} else {
			logger.file = file
		}
	}

	return logger
}

// parseLogLevel 解析日志级别
func parseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.level = level
}

// GetLevel 获取当前日志级别
func (l *Logger) GetLevel() LogLevel {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.level
}

// isLevelEnabled 检查日志级别是否启用
func (l *Logger) isLevelEnabled(level LogLevel) bool {
	return level >= l.GetLevel()
}

// log 核心日志方法
func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}) {
	if !l.isLevelEnabled(level) {
		return
	}

	// 获取调用者信息
	pc, file, line, ok := runtime.Caller(3)
	var function string
	if ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			function = fn.Name()
			// 简化函数名
			if idx := strings.LastIndex(function, "/"); idx != -1 {
				function = function[idx+1:]
			}
		}
		// 简化文件路径
		if idx := strings.LastIndex(file, "/"); idx != -1 {
			file = file[idx+1:]
		}
	}

	// 创建日志条目
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
		File:      file,
		Line:      line,
		Function:  function,
		Fields:    fields,
	}

	// 格式化输出
	var output string
	if l.format == "json" {
		if data, err := json.Marshal(entry); err == nil {
			output = string(data)
		} else {
			output = fmt.Sprintf("{\"level\":\"%s\",\"message\":\"%s\",\"error\":\"json marshal failed: %v\"}", level.String(), message, err)
		}
	} else {
		// 文本格式
		output = fmt.Sprintf("[%s] %s %s:%d %s - %s",
			entry.Timestamp, entry.Level, entry.File, entry.Line, entry.Function, entry.Message)
		if len(fields) > 0 {
			if fieldsJSON, err := json.Marshal(fields); err == nil {
				output += fmt.Sprintf(" fields=%s", string(fieldsJSON))
			}
		}
	}

	// 输出日志
	l.writeLog(output)
}

// writeLog 写入日志
func (l *Logger) writeLog(message string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.enableFile && l.file != nil {
		fmt.Fprintln(l.file, message)
	} else {
		switch l.output {
		case "stderr":
			fmt.Fprintln(os.Stderr, message)
		default: // stdout
			fmt.Fprintln(os.Stdout, message)
		}
	}
}

// Debug 调试日志
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(DEBUG, message, f)
}

// Info 信息日志
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(INFO, message, f)
}

// Infof 格式化信息日志
func (l *Logger) Infof(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(INFO, message, nil)
}

// Warn 警告日志
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(WARN, message, f)
}

// Warnf 格式化警告日志
func (l *Logger) Warnf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(WARN, message, nil)
}

// Error 错误日志
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ERROR, message, f)
}

// Errorf 格式化错误日志
func (l *Logger) Errorf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(ERROR, message, nil)
}

// Fatal 致命错误日志
func (l *Logger) Fatal(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(FATAL, message, f)
	os.Exit(1)
}

// WithFields 带字段的日志
func (l *Logger) WithFields(fields map[string]interface{}) *FieldLogger {
	return &FieldLogger{
		logger: l,
		fields: fields,
	}
}

// Close 关闭日志服务
func (l *Logger) Close() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}

// FieldLogger 带字段的日志器
type FieldLogger struct {
	logger *Logger
	fields map[string]interface{}
}

// Debug 调试日志
func (fl *FieldLogger) Debug(message string) {
	fl.logger.log(DEBUG, message, fl.fields)
}

// Info 信息日志
func (fl *FieldLogger) Info(message string) {
	fl.logger.log(INFO, message, fl.fields)
}

// Warn 警告日志
func (fl *FieldLogger) Warn(message string) {
	fl.logger.log(WARN, message, fl.fields)
}

// Error 错误日志
func (fl *FieldLogger) Error(message string) {
	fl.logger.log(ERROR, message, fl.fields)
}

// Fatal 致命错误日志
func (fl *FieldLogger) Fatal(message string) {
	fl.logger.log(FATAL, message, fl.fields)
	os.Exit(1)
}

// 全局便捷方法
var globalLogger *Logger
var globalLoggerMutex sync.RWMutex

// getGlobalLogger 获取全局日志器（支持动态更新）
func getGlobalLogger() *Logger {
	globalLoggerMutex.RLock()
	if globalLogger != nil {
		defer globalLoggerMutex.RUnlock()
		return globalLogger
	}
	globalLoggerMutex.RUnlock()

	// 如果全局日志器未设置，创建一个临时的DEBUG级别日志器
	globalLoggerMutex.Lock()
	defer globalLoggerMutex.Unlock()

	// 双重检查，防止并发创建
	if globalLogger == nil {
		globalLogger = &Logger{
			level:      DEBUG, // 改为DEBUG级别，确保所有日志都能显示
			format:     "json",
			output:     "stdout",
			enableFile: false,
		}
	}
	return globalLogger
}

// Debug 全局调试日志
func Debug(message string, fields ...map[string]interface{}) {
	getGlobalLogger().Debug(message, fields...)
}

// Info 全局信息日志
func Info(message string, fields ...map[string]interface{}) {
	getGlobalLogger().Info(message, fields...)
}

// Infof 全局格式化信息日志
func Infof(format string, args ...interface{}) {
	getGlobalLogger().Infof(format, args...)
}

// Warn 全局警告日志
func Warn(message string, fields ...map[string]interface{}) {
	getGlobalLogger().Warn(message, fields...)
}

// Warnf 全局格式化警告日志
func Warnf(format string, args ...interface{}) {
	getGlobalLogger().Warnf(format, args...)
}

// Error 全局错误日志
func Error(message string, fields ...map[string]interface{}) {
	getGlobalLogger().Error(message, fields...)
}

// Errorf 全局格式化错误日志
func Errorf(format string, args ...interface{}) {
	getGlobalLogger().Errorf(format, args...)
}

// Fatal 全局致命错误日志
func Fatal(message string, fields ...map[string]interface{}) {
	getGlobalLogger().Fatal(message, fields...)
}

// WithFields 全局带字段日志
func WithFields(fields map[string]interface{}) *FieldLogger {
	return getGlobalLogger().WithFields(fields)
}

// SetGlobalLogger 设置全局日志器（线程安全）
func SetGlobalLogger(logger *Logger) {
	globalLoggerMutex.Lock()
	defer globalLoggerMutex.Unlock()
	globalLogger = logger
}

// GetGlobalLogger 获取当前全局日志器（用于调试）
func GetGlobalLogger() *Logger {
	globalLoggerMutex.RLock()
	defer globalLoggerMutex.RUnlock()
	return globalLogger
}

// IsGlobalLoggerInitialized 检查全局日志器是否已初始化
func IsGlobalLoggerInitialized() bool {
	globalLoggerMutex.RLock()
	defer globalLoggerMutex.RUnlock()
	return globalLogger != nil
}
