package middleware

import (
	"time"

	"openbridge/backend/internal/pkg/logger"
	"openbridge/backend/internal/pkg/myerror"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AccessLog 是一个 Gin 中间件函数，用于记录 HTTP 请求的访问日志
func AccessLog() gin.HandlerFunc {
	// 返回一个匿名函数作为中间件处理函数
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()
		// 获取请求路径
		path := c.Request.URL.Path
		// 获取请求方法
		method := c.Request.Method
		// 获取请求ID
		requestID := GetRequestID(c)

		// 继续处理后续的中间件和路由处理函数
		c.Next()

		// 获取响应体
		status := c.GetInt(logger.LoggerErrorCodeKey)
		if status == 0 {
			status = myerror.ErrorCodeOK
		}

		// 计算请求处理耗时（秒）
		latency := time.Since(start).Seconds()

		// 获取日志消息
		msg := c.GetString(logger.LoggerMessageKey)

		// 使用 zap 记录访问日志
		logger.L().Info(msg,
			zap.String("request_id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Float64("latency", latency),
		)
	}
}
