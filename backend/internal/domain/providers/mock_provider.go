package providers

import (
	"context"
	"openbridge/backend/internal/domain/entity"
)

type MockProvider struct {
}

func (p *MockProvider) Name() string {
	return "mock"
}

func (p *MockProvider) GetQuota(ctx context.Context, account *entity.ProviderAccount) (entity.Quota, error) {
	return entity.Quota{
		Provider: "mock",
		Total: 1000, // 100GB
		Used:  200,
		Available: 800,
	}, nil
}

func (p *MockProvider) GetDirectLink(ctx context.Context, fileID string, account *entity.ProviderAccount) (string, error) {
	return "https://mockprovider.com/directlink/" + fileID, nil
}

func (p *MockProvider) RefreshToken(ctx context.Context, account *entity.ProviderAccount) error {
	return nil
}