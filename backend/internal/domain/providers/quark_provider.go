package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"openbridge/backend/internal/domain/entity"
	"openbridge/backend/internal/repository"
	"strings"
	"time"
)

const quarkQuotaURL = "https://drive-pc.quark.cn/1/clouddrive/member"

const quarkResponseOK = 0

const bytesPerMB = 1024 * 1024

type QuarkProvider struct {
	client       *http.Client
	providerRepo *repository.ProviderRepository
}

type quarkQuotaResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    quarkQuotaDetail `json:"data"`
}

type quarkQuotaDetail struct {
	Total int64 `json:"total_capacity"`
	Used  int64 `json:"use_capacity"`
	// Free  int64 `json:"free"`
}

func NewQuarkProvider(providerRepo *repository.ProviderRepository) *QuarkProvider {
	return &QuarkProvider{
		client:       &http.Client{Timeout: 10 * time.Second},
		providerRepo: providerRepo,
	}
}

func (p *QuarkProvider) Name() string {
	return "quark"
}

func (p *QuarkProvider) GetQuota(ctx context.Context, account *entity.ProviderAccount) (entity.Quota, error) {
	if account == nil {
		return entity.Quota{}, fmt.Errorf("quark provider: account is nil")
	}
	cookie := strings.TrimSpace(account.AuthCookie)
	if cookie == "" {
		return entity.Quota{}, fmt.Errorf("quark provider: cookie is empty")
	}

	// fmt.Printf("!!!!!!!!!!!!!quark provider: get quota with cookie=%s\n", cookie)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, quarkQuotaURL, nil)
	if err != nil {
		return entity.Quota{}, fmt.Errorf("quark provider: create request failed: %w", err)
	}
	q := req.URL.Query()
	q.Set("pr", "ucpro")
	q.Set("fr", "pc")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Referer", "https://pan.quark.cn/")

	resp, err := p.client.Do(req)
	if err != nil {
		return entity.Quota{}, fmt.Errorf("quark provider: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return entity.Quota{}, fmt.Errorf("quark provider: cookie invalid or expired, status=%d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return entity.Quota{}, fmt.Errorf("quark provider: unexpected status=%d", resp.StatusCode)
	}

	var payload quarkQuotaResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return entity.Quota{}, fmt.Errorf("quark provider: parse response failed: %w", err)
	}
	if payload.Code != quarkResponseOK {
		return entity.Quota{}, fmt.Errorf("quark provider: api code=%d message=%s", payload.Code, payload.Message)
	}
	if payload.Data.Total < 0 || payload.Data.Used < 0 || payload.Data.Used > payload.Data.Total {
		return entity.Quota{}, fmt.Errorf("quark provider: invalid quota fields total=%d used=%d", payload.Data.Total, payload.Data.Used)
	}

	available := payload.Data.Total - payload.Data.Used
	if available == 0 && payload.Data.Total >= payload.Data.Used {
		available = payload.Data.Total - payload.Data.Used
	}

	if payload.Data.Total > 0 {
		payload.Data.Total = payload.Data.Total / bytesPerMB
	}
	if payload.Data.Used > 0 {
		payload.Data.Used = payload.Data.Used / bytesPerMB
	}
	if available > 0 {
		available = available / bytesPerMB
	}

	now := time.Now().UTC()
	return entity.Quota{
		Provider:  "quark",
		Total:     payload.Data.Total,
		Used:      payload.Data.Used,
		Available: available,
		UpdatedAt: now,
	}, nil
}

func (p *QuarkProvider) GetDirectLink(ctx context.Context, fileID string, account *entity.ProviderAccount) (string, error) {
	_ = ctx
	_ = fileID
	_ = account
	return "", fmt.Errorf("quark provider: direct link not supported")
}

func (p *QuarkProvider) RefreshToken(ctx context.Context, account *entity.ProviderAccount) error {
	_ = ctx
	_ = account
	return fmt.Errorf("quark provider: refresh token not supported")
}
