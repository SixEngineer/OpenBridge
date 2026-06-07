package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var appLogger *zap.Logger

const (
	LoggerMessageKey     = "msg"
	LoggerErrorCodeKey   = "errorcode"
)

// Init initializes global logger based on level and format.
// Init 初始化日志记录器
// level: 日志级别，如 "debug", "info", "warn", "error"
// format: 日志输出格式，支持 "console" 或 "json"
// 返回值: 初始化过程中的错误信息
func Init(level, format string) error {
	// 设置默认日志级别为InfoLevel
	logLevel := zapcore.InfoLevel
	// 尝试将输入的日志级别字符串转换为zap的日志级别
	if err := logLevel.UnmarshalText([]byte(strings.ToLower(level))); err != nil {
		// 如果转换失败，则使用默认的InfoLevel
		logLevel = zapcore.InfoLevel
	}

	// 创建生产环境的日志配置
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(logLevel)

	// 设置时间戳格式为ISO8601
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 根据输入的格式设置日志格式
	if strings.EqualFold(format, "console") {
		cfg.Encoding = "console"
	} else {
		cfg.Encoding = "json"
	}

	// 使用配置创建日志记录器
	logger, err := cfg.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return err
	}

	// 将创建的日志记录器赋值给全局变量
	appLogger = logger
	return nil
}

// L 返回一个全局的 zap.Logger 实例
// 如果全局日志记录器尚未初始化，则会使用生产配置创建一个新的日志记录器
func L() *zap.Logger {
    // 检查全局日志记录器是否为 nil
	if appLogger == nil {
        // 如果未初始化，则使用 zap.NewProduction() 创建一个新的生产级日志记录器
        // 忽略可能的错误，因为这是回退方案
		fallback, _ := zap.NewProduction()
        // 将新创建的日志记录器赋值给全局变量 appLogger
		appLogger = fallback
	}
    // 返回全局日志记录器
	return appLogger
}

// Sync 函数用于同步日志记录器
// 如果 appLogger 不为 nil，则调用其 Sync 方法进行同步操作
func Sync() {
    // 检查 appLogger 是否已经初始化
	if appLogger != nil {
	    // 调用 appLogger 的 Sync 方法，使用下划线忽略可能的返回值
		_ = appLogger.Sync()
	}
}
