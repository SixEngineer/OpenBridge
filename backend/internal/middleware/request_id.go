package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const (
	RequestIDHeader = "X-Request-ID"
	RequestIDKey    = "request_id"
)

// RequestID 是一个 Gin 中间件函数，用于处理和设置请求 ID
// 它会检查请求头中是否已存在请求 ID，如果不存在则生成一个新的
// 然后将请求 ID 存储到上下文中，并添加到响应头中
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取请求 ID
		requestID := c.GetHeader(RequestIDHeader)
		// 如果请求头中没有请求 ID，则生成一个新的
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 将请求 ID 存储到上下文中，方便后续使用
		c.Set(RequestIDKey, requestID)
		// 将请求 ID 添加到响应头中，返回给客户端
		c.Writer.Header().Set(RequestIDHeader, requestID)
		// 继续处理后续的中间件和路由处理函数
		c.Next()
	}
}

// GetRequestID 从gin.Context中获取请求ID
// 参数:
//   c - gin.Context上下文对象，包含请求和响应的信息
// 返回值:
//   string - 返回请求ID字符串，如果获取失败则返回空字符串
func GetRequestID(c *gin.Context) string {
    // 尝试从上下文中获取请求ID
	v, ok := c.Get(RequestIDKey)
	if !ok {
        // 如果请求ID不存在，返回空字符串
		return ""
	}

    // 断言获取的值为字符串类型
	requestID, ok := v.(string)
	if !ok {
        // 如果类型断言失败，返回空字符串
		return ""
	}

    // 返回获取到的请求ID
	return requestID
}

// generateRequestID 生成一个唯一的请求ID
// 返回一个以"req_"开头的16进制字符串，如果生成失败则返回"req_fallback"
func generateRequestID() string {
	// 创建一个长度为12的字节数组
	b := make([]byte, 12)
	// 尝试从随机源读取随机数据
	if _, err := rand.Read(b); err != nil {
		// 如果读取失败，返回默认的请求ID
		return "req_fallback"
	}
	// 将随机字节数组转换为16进制字符串，并添加"req_"前缀返回
	return "req_" + hex.EncodeToString(b)
}
