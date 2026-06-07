package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"openbridge/backend/internal/domain/entity"
	"openbridge/backend/internal/repository"
	"strings"
	"time"
)

const baiduQuotaURL = "https://pan.baidu.com/api/quota?checkfree=1"
const baiduRefreshTokenURL = "https://openapi.baidu.com/oauth/2.0/token"
const baiduAccessTokenDefaultTTLSeconds = 2592000

const baiduAPPKEY = "Cas0bx01638g4YqUS1xiCtSOw5qolQUu"
const baiduSECRETKEY = "Ma7HAdBDYhlE1znXalB7u6vrHhh3L9U0"

type BaiduProvider struct {
	client       *http.Client
	providerRepo *repository.ProviderRepository
}

type baiduQuotaResponse struct {
	Errno      int    `json:"errno"`
	Errmsg     string `json:"errmsg"`
	Total      int64  `json:"total"`
	Used       int64  `json:"used"`
	RequestID  int64  `json:"request_id"`
	GUIDInfo   int64  `json:"guid_info"`
	Expire     bool   `json:"expire"`
	Free       int64  `json:"free"`
	ActiveType int    `json:"active_type"`
}

type baiduRefreshTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func NewBaiduProvider(providerRepo *repository.ProviderRepository) *BaiduProvider {
	return &BaiduProvider{
		client:       &http.Client{Timeout: 10 * time.Second},
		providerRepo: providerRepo,
	}
}

func (p *BaiduProvider) Name() string {
	return "baidu"
}

// GetQuota 是百度提供商获取配额信息的方法
// 它接收一个上下文和提供商账户信息，返回配额信息和可能的错误
func (p *BaiduProvider) GetQuota(ctx context.Context, account *entity.ProviderAccount) (entity.Quota, error) {
	// 检查账户是否为空
	if account == nil {
		return entity.Quota{}, fmt.Errorf("baidu provider: account is nil")
	}
	// 检查访问令牌是否为空
	if strings.TrimSpace(account.AccessToken) == "" {
		return entity.Quota{}, fmt.Errorf("baidu provider: access token is empty")
	}

	// 创建新的HTTP请求，使用GET方法访问百度配额URL
	baiduQuotaURLWithToken := fmt.Sprintf("%s&access_token=%s", baiduQuotaURL, account.AccessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baiduQuotaURLWithToken, nil)
	if err != nil {
		return entity.Quota{}, fmt.Errorf("baidu provider: create request failed: %w", err)
	}
	// 设置请求头
	req.Header.Set("User-Agent", "OpenBridge/1.0")
	// req.Header.Set("Authorization", "Bearer "+account.AccessToken)

	// 发送HTTP请求
	resp, err := p.client.Do(req)
	if err != nil {
		return entity.Quota{}, fmt.Errorf("baidu provider: request failed: %w", err)
	}
	defer resp.Body.Close() // 确保响应体被关闭

	// 检查响应状态码，处理未授权或禁止访问的情况
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return entity.Quota{}, fmt.Errorf("baidu provider: token invalid or expired, status=%d", resp.StatusCode)
	}
	// 检查响应状态码是否为200，如果不是则返回错误
	if resp.StatusCode != http.StatusOK {
		// 处理令牌无效或过期的情况，例如触发令牌刷新机制
		// err := p.RefreshToken(ctx, account)
		// if err != nil {
		// 	return entity.Quota{}, fmt.Errorf("baidu provider: refresh token failed: %w", err)
		// }
		return entity.Quota{}, fmt.Errorf("baidu provider: unexpected status=%d", resp.StatusCode)
	}

	// 解析响应体到百度配额响应结构体
	var payload baiduQuotaResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return entity.Quota{}, fmt.Errorf("baidu provider: parse response failed: %w", err)
	}

	// 检查API返回的错误码
	if payload.Errno != 0 {
		// 处理令牌无效或过期的情况
		if payload.Errno == -6 || payload.Errno == 111 {
			return entity.Quota{}, fmt.Errorf("baidu provider: token invalid or expired, errno=%d errmsg=%s", payload.Errno, payload.Errmsg)
		}
		// 处理其他API错误
		return entity.Quota{}, fmt.Errorf("baidu provider: api errno=%d errmsg=%s", payload.Errno, payload.Errmsg)
	}
	// 验证配额字段的合法性
	if payload.Total < 0 || payload.Used < 0 || payload.Used > payload.Total {
		return entity.Quota{}, fmt.Errorf("baidu provider: invalid quota fields total=%d used=%d", payload.Total, payload.Used)
	}

	// 创建当前UTC时间
	now := time.Now().UTC()

	// 转换为 MB
	payload.Total /= (1024 * 1024)
	payload.Used /= (1024 * 1024)

	// 返回格式化后的配额信息
	return entity.Quota{
		Provider:  "baidu",
		Total:     payload.Total,
		Used:      payload.Used,
		Available: payload.Total - payload.Used,
		UpdatedAt: now,
	}, nil
}

func (p *BaiduProvider) GetDirectLink(ctx context.Context, fileID string, account *entity.ProviderAccount) (string, error) {
	// TODO: 实现获取百度网盘文件的直接链接的逻辑
	return "", fmt.Errorf("baidu provider: direct link not implemented")
}


// ！！！ 注意：此方法暂时无法使用，因为基于百度网盘的应用还没有在百度开放平台提交审核
// ！！！
// ！！！
func (p *BaiduProvider) RefreshToken(ctx context.Context, account *entity.ProviderAccount) error {

	if account == nil {
		return fmt.Errorf("baidu provider: account is nil")
	}
	refreshToken := strings.TrimSpace(account.RefreshToken)
	if refreshToken == "" {
		return fmt.Errorf("baidu provider: refresh token is empty")
	}
	if p.providerRepo == nil {
		return fmt.Errorf("baidu provider: provider repository is nil")
	}

	params := url.Values{}
	params.Set("grant_type", "refresh_token")
	params.Set("refresh_token", refreshToken)
	params.Set("client_id", baiduAPPKEY)
	params.Set("client_secret", baiduSECRETKEY)

	refreshURL := fmt.Sprintf("%s?%s", baiduRefreshTokenURL, params.Encode())
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, refreshURL, nil)
	if err != nil {
		return fmt.Errorf("baidu provider: create refresh token request failed: %w", err)
	}
	request.Header.Set("User-Agent", "OpenBridge/1.0")

	resp, err := p.client.Do(request)
	if err != nil {
		return fmt.Errorf("baidu provider: refresh token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("baidu provider: refresh token status=%d", resp.StatusCode)
	}

	var payload baiduRefreshTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("baidu provider: refresh token parse failed: %w", err)
	}
	if strings.TrimSpace(payload.Error) != "" {
		return fmt.Errorf("baidu provider: refresh token api error=%s description=%s", payload.Error, payload.ErrorDescription)
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return fmt.Errorf("baidu provider: refresh token response missing access_token")
	}

	expiresIn := payload.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = baiduAccessTokenDefaultTTLSeconds
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(expiresIn) * time.Second)

	account.AccessToken = payload.AccessToken
	if strings.TrimSpace(payload.RefreshToken) != "" {
		account.RefreshToken = payload.RefreshToken
	}
	account.TokenExpiresAt = &expiresAt

	if account.ID == 0 {
		return fmt.Errorf("baidu provider: account id is empty")
	}

	if err := p.providerRepo.UpdateProviderAccount(account); err != nil {
		return fmt.Errorf("baidu provider: update provider account failed: %w", err)
	}

	return nil
}
