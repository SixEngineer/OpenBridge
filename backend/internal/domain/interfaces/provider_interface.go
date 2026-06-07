package interfaces

import (
	"context"
	"openbridge/backend/internal/domain/entity"
)

type Provider interface {
	Name() string
	GetQuota(ctx context.Context, account *entity.ProviderAccount) (entity.Quota, error)
	GetDirectLink(ctx context.Context, fileID string, account *entity.ProviderAccount) (string, error)
	RefreshToken(ctx context.Context, account *entity.ProviderAccount) error
}