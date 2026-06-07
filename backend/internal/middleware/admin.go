package middleware

import (
	"encoding/json"
	"net/http"
	"openbridge/backend/internal/pkg/myerror"
	"openbridge/backend/internal/pkg/logger"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminChecker 通过 OpenList /api/me 验证当前用户是否为管理员（role == 0）
// 内置 60 秒缓存，避免每个请求都调用 OpenList API
type AdminChecker struct {
	baseURL    string
	token      string
	mu         sync.Mutex
	lastCheck  time.Time
	cachedRole int
	cacheTTL   time.Duration
	lastToken  string // 记录缓存的 token，token 变化时强制重新检查
}

func NewAdminChecker(baseURL string) *AdminChecker {
	return &AdminChecker{
		baseURL:  baseURL,
		cacheTTL: 60 * time.Second,
	}
}

// SetToken 更新 token（由登录后调用），同时清除缓存
func (c *AdminChecker) SetToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
	c.lastCheck = time.Time{} // 清空缓存，下次强制重新检查
}

// InvalidateCache 清除缓存，下次请求强制重新检查
func (c *AdminChecker) InvalidateCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastCheck = time.Time{}
}

// IsAdmin 检查当前用户是否为管理员
func (c *AdminChecker) IsAdmin() bool {
	role, err := c.fetchRole()
	if err != nil {
		logger.L().Warn("admin check failed", zap.Error(err))
		return false
	}
	// OpenList 角色: 0=GENERAL, 1=GUEST, 2=ADMIN
	return role == 2
}

// fetchRole 获取当前用户的角色（带缓存）
func (c *AdminChecker) fetchRole() (int, error) {
	c.mu.Lock()
	// 缓存命中且 token 未变
	tokenChanged := c.lastToken != c.token
	if time.Since(c.lastCheck) < c.cacheTTL && c.lastCheck.After(time.Time{}) && !tokenChanged {
		role := c.cachedRole
		c.mu.Unlock()
		return role, nil
	}
	c.mu.Unlock()

	// 调用 OpenList API
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", c.baseURL+"/api/me", nil)
	if err != nil {
		return -1, err
	}
	req.Header.Set("Content-Type", "application/json")

	c.mu.Lock()
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}
	c.mu.Unlock()

	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			Role int `json:"role"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return -1, err
	}

	// 更新缓存
	c.mu.Lock()
	c.cachedRole = result.Data.Role
	c.lastCheck = time.Now()
	c.lastToken = c.token
	c.mu.Unlock()

	return result.Data.Role, nil
}

// Middleware 返回 Gin 中间件，非管理员返回 403
func (c *AdminChecker) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !c.IsAdmin() {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    myerror.ErrorCodeForbidden,
				"message": "admin only",
			})
			return
		}
		ctx.Next()
	}
}
