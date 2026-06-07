//go:build windows
// +build windows

package providers

import (
	"context"
	"fmt"
	"openbridge/backend/internal/domain/entity"
	"openbridge/backend/internal/repository"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

type LocalWindowsProvider struct {
	providerRepo *repository.ProviderRepository
	mountRepo    *repository.MountRepository
}

func NewLocalProvider(providerRepo *repository.ProviderRepository, mountRepo *repository.MountRepository) *LocalWindowsProvider {
	return &LocalWindowsProvider{
		providerRepo: providerRepo,
		mountRepo:    mountRepo,
	}
}

func (p *LocalWindowsProvider) Name() string {
	return "local_windows"
}

func (p *LocalWindowsProvider) GetQuota(ctx context.Context, account *entity.ProviderAccount) (entity.Quota, error) {
	_ = ctx
	if account == nil {
		return entity.Quota{}, fmt.Errorf("local windows provider: account is nil")
	}

	mountPoint, err := p.mountRepo.GetMountPointByProviderAccountID(account.ID)
	if err != nil {
		return entity.Quota{}, fmt.Errorf("local windows provider: get mount point failed: %w", err)
	}
	path := mountPoint.ProviderRootPath

	path = strings.ReplaceAll(path, "/", "\\")
	if len(path) == 2 && path[1] == ':' {
		path = path + "\\"
	}
	path = filepath.Clean(path)
	if !strings.HasSuffix(path, "\\") {
		path = path + "\\"
	}

	var freeBytesAvailable uint64
	var totalBytes uint64
	var totalFreeBytes uint64
	if err := windows.GetDiskFreeSpaceEx(windows.StringToUTF16Ptr(path), &freeBytesAvailable, &totalBytes, &totalFreeBytes); err != nil {
		return entity.Quota{}, fmt.Errorf("local windows provider: get disk space failed for %s: %w", path, err)
	}

	if totalBytes == 0 || totalFreeBytes > totalBytes {
		return entity.Quota{}, fmt.Errorf("local windows provider: invalid disk stats total=%d free=%d", totalBytes, totalFreeBytes)
	}

	usedBytes := int64(totalBytes - totalFreeBytes)
	totalMB := int64(totalBytes / (1024 * 1024))
	usedMB := int64(usedBytes / (1024 * 1024))
	availableMB := int64(totalFreeBytes / (1024 * 1024))

	now := time.Now().UTC()
	return entity.Quota{
		Provider:  "local_windows",
		Total:     totalMB,
		Used:      usedMB,
		Available: availableMB,
		UpdatedAt: now,
	}, nil
}

func (p *LocalWindowsProvider) GetDirectLink(ctx context.Context, fileID string, account *entity.ProviderAccount) (string, error) {
	_ = ctx
	_ = fileID
	_ = account
	return "", fmt.Errorf("local windows provider: direct link not supported")
}

func (p *LocalWindowsProvider) RefreshToken(ctx context.Context, account *entity.ProviderAccount) error {
	_ = ctx
	_ = account
	return fmt.Errorf("local windows provider: refresh token not supported")
}
