//go:build linux
// +build linux

package providers

import (
	"context"
	"fmt"
	"openbridge/backend/internal/domain/entity"
	"openbridge/backend/internal/repository"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// LocalLinuxProvider Linux本地存储Provider
type LocalLinuxProvider struct {
	providerRepo *repository.ProviderRepository
	mountRepo    *repository.MountRepository
}

// NewLocalLinuxProvider 创建Linux本地存储Provider实例
func NewLocalProvider(providerRepo *repository.ProviderRepository, mountRepo *repository.MountRepository) *LocalLinuxProvider {
	return &LocalLinuxProvider{
		providerRepo: providerRepo,
		mountRepo:    mountRepo,
	}
}

// Name 返回Provider名称
func (p *LocalLinuxProvider) Name() string {
	return "local_linux"
}

// GetQuota 获取磁盘配额信息
func (p *LocalLinuxProvider) GetQuota(ctx context.Context, account *entity.ProviderAccount) (entity.Quota, error) {
	_ = ctx
	if account == nil {
		return entity.Quota{}, fmt.Errorf("local linux provider: account is nil")
	}

	mountPoint, err := p.mountRepo.GetMountPointByProviderAccountID(account.ID)
	if err != nil {
		return entity.Quota{}, fmt.Errorf("local linux provider: get mount point failed: %w", err)
	}

	path := mountPoint.ProviderRootPath

	// 路径标准化处理
	// 1. 替换Windows风格的反斜杠为正斜杠
	path = strings.ReplaceAll(path, "\\", "/")
	// 2. 清理路径，移除冗余的分隔符和.、..等
	path = filepath.Clean(path)
	// 3. 确保路径以/开头（绝对路径）
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Linux系统调用获取磁盘空间信息
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return entity.Quota{}, fmt.Errorf("local linux provider: get disk space failed for %s: %w", path, err)
	}

	// 计算磁盘空间（字节）
	// stat.Blocks: 文件系统总块数
	// stat.Bsize: 每块大小（字节）
	totalBytes := stat.Blocks * uint64(stat.Bsize)

	// stat.Bfree: 空闲块数（包括保留块）
	freeBytes := stat.Bfree * uint64(stat.Bsize)

	// stat.Bavail: 非特权用户可用的空闲块数
	availableBytes := stat.Bavail * uint64(stat.Bsize)

	// 已使用字节数
	usedBytes := totalBytes - freeBytes

	// 数据验证
	if totalBytes == 0 || freeBytes > totalBytes {
		return entity.Quota{}, fmt.Errorf("local linux provider: invalid disk stats total=%d free=%d", totalBytes, freeBytes)
	}

	// 单位转换：字节 -> MB
	totalMB := int64(totalBytes / (1024 * 1024))
	usedMB := int64(usedBytes / (1024 * 1024))
	availableMB := int64(availableBytes / (1024 * 1024))

	// 返回配额信息
	now := time.Now().UTC()
	return entity.Quota{
		Provider:  "local_linux",
		Total:     totalMB,
		Used:      usedMB,
		Available: availableMB,
		UpdatedAt: now,
	}, nil
}

// GetDirectLink 获取文件直链（Linux本地存储不支持）
func (p *LocalLinuxProvider) GetDirectLink(ctx context.Context, fileID string, account *entity.ProviderAccount) (string, error) {
	_ = ctx
	_ = fileID
	_ = account
	return "", fmt.Errorf("local linux provider: direct link not supported")
}

// RefreshToken 刷新令牌（Linux本地存储不支持）
func (p *LocalLinuxProvider) RefreshToken(ctx context.Context, account *entity.ProviderAccount) error {
	_ = ctx
	_ = account
	return fmt.Errorf("local linux provider: refresh token not supported")
}
