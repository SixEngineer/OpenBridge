package usecase

import (
	"context"
	"errors"
	"fmt"
	"openbridge/backend/internal/config"
	"openbridge/backend/internal/domain/entity"
	"openbridge/backend/internal/domain/interfaces"
	"openbridge/backend/internal/domain/providers"
	"openbridge/backend/internal/pkg/logger"
	"openbridge/backend/internal/repository"
	"openbridge/backend/internal/tool"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrMountInvalidMode           = errors.New("mount quota_mode is invalid")
	ErrMountPathRequired          = errors.New("mount path is required")
	ErrMountProviderRequired      = errors.New("real mode requires provider_account_id")
	ErrMountParentRequired        = errors.New("inherit mode requires inherit_from_id")
	ErrMountParentNotReal         = errors.New("inherit parent must be real mode")
	ErrMountCircularInherit       = errors.New("inherit chain has cycle")
	ErrMountVirtualExceedsAllowed = errors.New("virtual_total exceeds allowed_max")
	ErrMountVirtualUsedInvalid    = errors.New("virtual_used must be <= virtual_total")
	ErrMountDisabled              = errors.New("mount is disabled")
	ErrProviderNotFound           = errors.New("provider not found")
	ErrMountDeleteInherited       = errors.New("mount is inherited by another mount, remove it first")
)

type MountQuotaResult struct {
	MountID       uint             `json:"mount_id"`
	Mode          string           `json:"mode"`
	AllowedMax    int64            `json:"allowed_max"`
	Quota         entity.Quota     `json:"quota"`
	InheritChain  []uint           `json:"inherit_chain,omitempty"`
	VirtualConfig map[string]int64 `json:"virtual_config,omitempty"`
}

type MountUseCase struct {
	mountRepo        *repository.MountRepository
	providerRepo     *repository.ProviderRepository
	quotaRepo        *repository.QuotaRepository
	providerRegistry *tool.Registry
	config           *config.Config
	profileSyncer    MountProfileSyncer
}

type MountProfileSyncer interface {
	HandleMountDeleted(mountID uint) error
}

func NewMountUseCase(mountRepo *repository.MountRepository, providerRepo *repository.ProviderRepository, quotaRepo *repository.QuotaRepository, providerRegistry *tool.Registry, cfg *config.Config) *MountUseCase {
	return &MountUseCase{
		mountRepo:        mountRepo,
		providerRepo:     providerRepo,
		quotaRepo:        quotaRepo,
		providerRegistry: providerRegistry,
		config:           cfg,
	}
}

func (u *MountUseCase) SetProfileSyncer(syncer MountProfileSyncer) {
	u.profileSyncer = syncer
}

// CreateMount 创建一个新的挂载点
func (u *MountUseCase) CreateMount(ctx context.Context, mount entity.MountPoint) (*entity.MountPoint, error) {
	mode := entity.QuotaMode(strings.ToLower(strings.TrimSpace(mount.QuotaMode)))
	mount.QuotaMode = string(mode)
	mount.MountPath = normalizeOpenListMountPath(mount.MountPath)
	if mount.MountPath == "" {
		return nil, ErrMountPathRequired
	}

	if err := u.validateMountConfig(ctx, &mount, mode); err != nil {
		return nil, err
	}

	if mode == entity.QuotaModeVirtual && mount.VirtualTotal == 0 {
		mount.ReadOnly = true
	}

	if err := u.mountRepo.InsertMountPoint(&mount); err != nil {
		return nil, err
	}
	return &mount, nil
}

// ListMountsByProvider 查询指定 provider 的所有挂载点
func (u *MountUseCase) ListMountsByProvider(ctx context.Context, providerAccountID uint) ([]entity.MountPoint, error) {
	return u.mountRepo.ListMountPointsByProviderAccountID(providerAccountID)
}

// ListAllMounts 查询所有挂载点
func (u *MountUseCase) ListAllMounts(ctx context.Context) ([]entity.MountPoint, error) {
	accounts, err := u.listScopedProviders()
	if err != nil {
		return nil, err
	}

	return u.mountRepo.ListMountPointsByProviderAccountIDs(providerIDs(accounts))
}

// UpdateMount 更新挂载点（支持修改名称、配额模式及模式关联字段）
func (u *MountUseCase) UpdateMount(ctx context.Context, mount entity.MountPoint) (*entity.MountPoint, error) {
	existing, err := u.mountRepo.GetMountPoint(mount.ID)
	if err != nil {
		return nil, err
	}
	if err := u.ensureProviderInCurrentScope(existing.ProviderAccountID); err != nil {
		return nil, err
	}

	if mount.Name != "" {
		existing.Name = mount.Name
	}
	if mount.MountPath != "" {
		normalizedMountPath := normalizeOpenListMountPath(mount.MountPath)
		if normalizedMountPath == "" {
			return nil, ErrMountPathRequired
		}
		existing.MountPath = normalizedMountPath
	}

	// 检查是否要修改配额模式
	newMode := strings.TrimSpace(mount.QuotaMode)
	if newMode != "" && strings.ToLower(newMode) != strings.ToLower(existing.QuotaMode) {
		// 切换到新模式，用 validateMountConfig 校验并清理不相关字段
		mode := entity.QuotaMode(strings.ToLower(newMode))
		if err := mode.Valid(); err != nil {
			return nil, err
		}
		// 构造一个临时 MountPoint 用于校验
		tmp := *existing
		tmp.QuotaMode = string(mode)
		tmp.InheritFromID = mount.InheritFromID
		tmp.VirtualTotal = mount.VirtualTotal
		tmp.VirtualUsed = mount.VirtualUsed
		if err := u.validateMountConfig(ctx, &tmp, mode); err != nil {
			return nil, err
		}
		// 应用新模式及清理后的字段
		existing.QuotaMode = string(mode)
		existing.InheritFromID = tmp.InheritFromID
		existing.VirtualTotal = tmp.VirtualTotal
		existing.VirtualUsed = tmp.VirtualUsed
	} else if strings.ToLower(existing.QuotaMode) == string(entity.QuotaModeVirtual) {
		// 同模式虚拟更新
		if mount.VirtualTotal > 0 {
			existing.VirtualTotal = mount.VirtualTotal
		}
		existing.VirtualUsed = mount.VirtualUsed
		if existing.VirtualUsed > existing.VirtualTotal {
			return nil, ErrMountVirtualUsedInvalid
		}
	} else if strings.ToLower(existing.QuotaMode) == string(entity.QuotaModeInherit) {
		// 同模式继承可更新父挂载点
		if mount.InheritFromID != nil {
			tmp := *existing
			tmp.InheritFromID = mount.InheritFromID
			if err := u.validateMountConfig(ctx, &tmp, entity.QuotaModeInherit); err != nil {
				return nil, err
			}
			existing.InheritFromID = mount.InheritFromID
		}
	}

	if err := u.mountRepo.UpdateMountPoint(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// DeleteMount 删除挂载点
func (u *MountUseCase) DeleteMount(ctx context.Context, mountID uint) error {
	mount, err := u.mountRepo.GetMountPoint(mountID)
	if err != nil {
		return err
	}
	if err := u.ensureProviderInCurrentScope(mount.ProviderAccountID); err != nil {
		return err
	}

	// 检查是否有其他挂载点继承自该挂载点
	mounts, err := u.ListAllMounts(ctx)
	if err != nil {
		return err
	}
	for _, m := range mounts {
		if m.InheritFromID != nil && *m.InheritFromID == mountID {
			return ErrMountDeleteInherited
		}
	}
	if err := u.mountRepo.DeleteMountPoint(mountID); err != nil {
		return err
	}
	if u.profileSyncer != nil {
		if err := u.profileSyncer.HandleMountDeleted(mountID); err != nil {
			return err
		}
	}
	return nil
}

func (u *MountUseCase) GetMountQuota(ctx context.Context, mountID uint) (MountQuotaResult, error) {
	return u.resolveMountQuota(ctx, mountID, false)
}

func (u *MountUseCase) SyncMountQuota(ctx context.Context, mountID uint) (MountQuotaResult, error) {
	return u.resolveMountQuota(ctx, mountID, true)
}

func (u *MountUseCase) GetMountForWebDAV(ctx context.Context, mountID uint) (*entity.MountPoint, error) {
	return u.getActiveScopedMount(mountID)
}

func (u *MountUseCase) GetMountQuotaReadonly(ctx context.Context, mountID uint) (MountQuotaResult, error) {
	mount, err := u.getActiveScopedMount(mountID)
	if err != nil {
		return MountQuotaResult{}, err
	}
	return u.GetMountQuotaReadonlyForMount(ctx, mount)
}

func (u *MountUseCase) GetMountQuotaReadonlyForMount(ctx context.Context, mount *entity.MountPoint) (MountQuotaResult, error) {
	if mount == nil {
		return MountQuotaResult{}, gorm.ErrRecordNotFound
	}
	return u.resolveByMode(ctx, mount, false, map[uint]struct{}{})
}

// resolveMountQuota 是一个用于解析挂载点配额的方法
func (u *MountUseCase) resolveMountQuota(ctx context.Context, mountID uint, syncRemote bool) (MountQuotaResult, error) {
	mount, err := u.getActiveScopedMount(mountID)
	if err != nil {
		return MountQuotaResult{}, err
	}

	result, err := u.resolveByMode(ctx, mount, syncRemote, map[uint]struct{}{})
	if err != nil {
		logger.L().Error("mount quota resolve failed",
			zap.Uint("mount_id", mount.ID),
			zap.String("mode", mount.QuotaMode),
			zap.Error(err),
		)
		now := time.Now().UTC()
		_ = u.quotaRepo.InsertQuotaSnapshot(&entity.QuotaSnapshot{
			MountPointID:      mount.ID,
			ProviderAccountID: mount.ProviderAccountID,
			Provider:          mount.ProviderType,
			Mode:              mount.QuotaMode,
			SyncStatus:        "failed",
			ErrorMessage:      err.Error(),
			SyncedAt:          now,
		})
		return MountQuotaResult{}, err
	}

	now := time.Now().UTC()
	if err := u.quotaRepo.InsertQuotaSnapshot(&entity.QuotaSnapshot{
		MountPointID:      mount.ID,
		ProviderAccountID: mount.ProviderAccountID,
		Provider:          mount.ProviderType,
		Mode:              mount.QuotaMode,
		Total:             result.Quota.Total,
		Used:              result.Quota.Used,
		Available:         result.Quota.Available,
		SyncStatus:        "success",
		SyncedAt:          now,
	}); err != nil {
		return MountQuotaResult{}, err
	}

	result.Quota.UpdatedAt = now
	logger.L().Info("mount quota resolved",
		zap.Uint("mount_id", mount.ID),
		zap.String("mode", mount.QuotaMode),
		zap.Int64("total", result.Quota.Total),
		zap.Int64("used", result.Quota.Used),
		zap.Int64("available", result.Quota.Available),
	)
	return result, nil
}

func (u *MountUseCase) getActiveScopedMount(mountID uint) (*entity.MountPoint, error) {
	mount, err := u.mountRepo.GetMountPoint(mountID)
	if err != nil {
		return nil, err
	}
	if err := u.ensureProviderInCurrentScope(mount.ProviderAccountID); err != nil {
		return nil, err
	}
	if !mount.Enabled {
		return nil, ErrMountDisabled
	}
	return mount, nil
}

// resolveByMode 根据挂载点的配额模式递归解析最终配额结果。
func (u *MountUseCase) resolveByMode(ctx context.Context, mount *entity.MountPoint, syncRemote bool, visited map[uint]struct{}) (MountQuotaResult, error) {
	if _, ok := visited[mount.ID]; ok {
		return MountQuotaResult{}, ErrMountCircularInherit
	}
	visited[mount.ID] = struct{}{}

	mode := entity.QuotaMode(strings.ToLower(mount.QuotaMode))
	switch mode {
	case entity.QuotaModeReal:
		quota, err := u.resolveRealQuota(ctx, mount, syncRemote)
		if err != nil {
			return MountQuotaResult{}, err
		}
		logger.L().Info("quota resolver mode",
			zap.String("mode", string(mode)),
			zap.Uint("mount_id", mount.ID),
		)
		return MountQuotaResult{
			MountID:    mount.ID,
			Mode:       mount.QuotaMode,
			AllowedMax: quota.Total,
			Quota:      quota,
		}, nil
	case entity.QuotaModeInherit:
		if mount.InheritFromID == nil {
			return MountQuotaResult{}, ErrMountParentRequired
		}
		parent, err := u.mountRepo.GetMountPoint(*mount.InheritFromID)
		if err != nil {
			return MountQuotaResult{}, err
		}
		if strings.ToLower(parent.QuotaMode) != string(entity.QuotaModeReal) {
			return MountQuotaResult{}, ErrMountParentNotReal
		}
		parentResult, err := u.resolveByMode(ctx, parent, syncRemote, visited)
		if err != nil {
			return MountQuotaResult{}, err
		}
		chain := []uint{parent.ID}
		logger.L().Info("quota resolver inherit chain",
			zap.Uint("mount_id", mount.ID),
			zap.Uint("parent_mount_id", parent.ID),
		)
		return MountQuotaResult{
			MountID:      mount.ID,
			Mode:         mount.QuotaMode,
			AllowedMax:   parentResult.AllowedMax,
			Quota:        parentResult.Quota,
			InheritChain: chain,
		}, nil
	case entity.QuotaModeVirtual:
		if mount.VirtualUsed > mount.VirtualTotal {
			return MountQuotaResult{}, ErrMountVirtualUsedInvalid
		}
		allowedMax, err := u.getAllowedMax(ctx, mount, syncRemote)
		if err != nil {
			logger.L().Warn("virtual allowed_max unavailable, fallback to virtual_total",
				zap.Uint("mount_id", mount.ID),
				zap.Error(err),
			)
			allowedMax = mount.VirtualTotal
		} else if mount.VirtualTotal > allowedMax {
			logger.L().Warn("virtual_total exceeds allowed_max",
				zap.Uint("mount_id", mount.ID),
				zap.Int64("virtual_total", mount.VirtualTotal),
				zap.Int64("allowed_max", allowedMax),
			)
		}
		logger.L().Info("quota resolver mode",
			zap.String("mode", string(mode)),
			zap.Uint("mount_id", mount.ID),
			zap.Int64("virtual_total", mount.VirtualTotal),
			zap.Int64("virtual_used", mount.VirtualUsed),
			zap.Int64("allowed_max", allowedMax),
		)
		return MountQuotaResult{
			MountID:    mount.ID,
			Mode:       mount.QuotaMode,
			AllowedMax: allowedMax,
			Quota: entity.Quota{
				Provider:  mount.ProviderType,
				Total:     mount.VirtualTotal,
				Used:      mount.VirtualUsed,
				Available: mount.VirtualTotal - mount.VirtualUsed,
			},
			VirtualConfig: map[string]int64{
				"virtual_total": mount.VirtualTotal,
				"virtual_used":  mount.VirtualUsed,
			},
		}, nil
	default:
		return MountQuotaResult{}, ErrMountInvalidMode
	}
}

// resolveRealQuota 用于解析和获取真实的配额信息
func (u *MountUseCase) resolveRealQuota(ctx context.Context, mount *entity.MountPoint, syncRemote bool) (entity.Quota, error) {
	if mount.ProviderAccountID == 0 {
		return entity.Quota{}, ErrMountProviderRequired
	}
	account, err := u.getScopedProviderAccount(mount.ProviderAccountID)
	if err != nil {
		return entity.Quota{}, err
	}

	if !syncRemote {
		if account.TotalQuota < account.UsedQuota || account.TotalQuota < 0 || account.UsedQuota < 0 {
			return entity.Quota{}, fmt.Errorf("invalid stored quota in provider account")
		}
		return entity.Quota{
			Provider:  mount.ProviderType,
			Total:     account.TotalQuota,
			Used:      account.UsedQuota,
			Available: account.AvailableQuota,
			UpdatedAt: account.UpdatedAt.UTC(),
		}, nil
	}

	providerInstance, err := u.resolveProvider(account)
	if err != nil {
		return entity.Quota{}, err
	}
	remoteQuota, err := providerInstance.GetQuota(ctx, account)
	if err != nil {
		return entity.Quota{}, err
	}
	logger.L().Info("provider quota fetched",
		zap.String("provider", account.NetDisk),
		zap.Uint("provider_account_id", account.ID),
		zap.Int64("total", remoteQuota.Total),
		zap.Int64("used", remoteQuota.Used),
		zap.Int64("available", remoteQuota.Available),
	)

	now := time.Now().UTC()
	if err := u.providerRepo.UpdateProviderQuota(account.ID, remoteQuota.Total, remoteQuota.Used, remoteQuota.Available, now); err != nil {
		return entity.Quota{}, err
	}
	remoteQuota.Provider = mount.ProviderType
	remoteQuota.UpdatedAt = now
	return remoteQuota, nil
}

// getAllowedMax 获取允许的最大挂载配额
func (u *MountUseCase) getAllowedMax(ctx context.Context, mount *entity.MountPoint, syncRemote bool) (int64, error) {
	account, err := u.getScopedProviderAccount(mount.ProviderAccountID)
	if err != nil {
		return 0, err
	}
	if !syncRemote && account.TotalQuota > 0 {
		return account.TotalQuota, nil
	}

	providerInstance, err := u.resolveProvider(account)
	if err != nil {
		return 0, err
	}
	quota, err := providerInstance.GetQuota(ctx, account)
	if err != nil {
		return 0, err
	}
	now := time.Now().UTC()
	if err := u.providerRepo.UpdateProviderQuota(account.ID, quota.Total, quota.Used, quota.Available, now); err != nil {
		return 0, err
	}
	logger.L().Info("virtual allowed_max evaluated",
		zap.Uint("mount_id", mount.ID),
		zap.Int64("allowed_max", quota.Total),
	)
	return quota.Total, nil
}

// validateMountConfig 验证挂载配置
func (u *MountUseCase) validateMountConfig(ctx context.Context, mount *entity.MountPoint, mode entity.QuotaMode) error {
	switch mode {
	case entity.QuotaModeReal:
		if mount.ProviderAccountID == 0 {
			return ErrMountProviderRequired
		}
		account, err := u.getScopedProviderAccount(mount.ProviderAccountID)
		if err != nil {
			return err
		}
		mount.ProviderType = account.NetDisk
		mount.InheritFromID = nil
		mount.VirtualTotal = 0
		mount.VirtualUsed = 0
		return nil
	case entity.QuotaModeInherit:
		if mount.InheritFromID == nil {
			return ErrMountParentRequired
		}
		parent, err := u.mountRepo.GetMountPoint(*mount.InheritFromID)
		if err != nil {
			return err
		}
		if err := u.ensureProviderInCurrentScope(parent.ProviderAccountID); err != nil {
			return err
		}
		if strings.ToLower(parent.QuotaMode) != string(entity.QuotaModeReal) {
			return ErrMountParentNotReal
		}
		if err := u.validateNoCycle(parent.ID, mount.ID); err != nil {
			return err
		}
		mount.VirtualTotal = 0
		mount.VirtualUsed = 0
		return nil
	case entity.QuotaModeVirtual:
		if mount.ProviderAccountID == 0 {
			return ErrMountProviderRequired
		}
		if mount.VirtualUsed > mount.VirtualTotal {
			return ErrMountVirtualUsedInvalid
		}
		account, err := u.getScopedProviderAccount(mount.ProviderAccountID)
		if err != nil {
			return err
		}
		mount.ProviderType = account.NetDisk
		return nil
	default:
		return ErrMountInvalidMode
	}
}

// validateNoCycle 检查挂载点是否存在继承循环
func (u *MountUseCase) validateNoCycle(startID uint, candidateID uint) error {
	visited := map[uint]struct{}{}
	currentID := startID
	for currentID != 0 {
		if currentID == candidateID {
			return ErrMountCircularInherit
		}
		if _, ok := visited[currentID]; ok {
			return ErrMountCircularInherit
		}
		visited[currentID] = struct{}{}
		current, err := u.mountRepo.GetMountPoint(currentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if err := u.ensureProviderInCurrentScope(current.ProviderAccountID); err != nil {
			return err
		}
		if current.InheritFromID == nil {
			return nil
		}
		currentID = *current.InheritFromID
	}
	return nil
}

func (u *MountUseCase) resolveProvider(account *entity.ProviderAccount) (interfaces.Provider, error) {
	key := providerRegistryKey(account)
	if providerInstance, ok := u.providerRegistry.Get(key); ok {
		return providerInstance, nil
	}

	providerInstance := buildMountProviderByNetDisk(account.NetDisk, u.providerRepo, u.mountRepo)
	if providerInstance == nil {
		return nil, ErrProviderNotFound
	}
	_ = u.providerRegistry.Register(key, providerInstance)
	return providerInstance, nil
}

func (u *MountUseCase) currentOpenListScope() string {
	return config.NormalizeBaseURLScope(u.config.OpenList.BaseURL)
}

func (u *MountUseCase) listScopedProviders() ([]entity.ProviderAccount, error) {
	scope := u.currentOpenListScope()
	if err := u.providerRepo.AssignEmptyOpenListScope(scope); err != nil {
		return nil, err
	}
	return u.providerRepo.ListProviderAccountsByOpenList(scope)
}

func (u *MountUseCase) getScopedProviderAccount(providerAccountID uint) (*entity.ProviderAccount, error) {
	scope := u.currentOpenListScope()
	if err := u.providerRepo.AssignEmptyOpenListScope(scope); err != nil {
		return nil, err
	}
	return u.providerRepo.GetProviderAccountByOpenList(providerAccountID, scope)
}

func (u *MountUseCase) ensureProviderInCurrentScope(providerAccountID uint) error {
	_, err := u.getScopedProviderAccount(providerAccountID)
	return err
}

func providerIDs(accounts []entity.ProviderAccount) []uint {
	ids := make([]uint, 0, len(accounts))
	for _, account := range accounts {
		ids = append(ids, account.ID)
	}
	return ids
}

func normalizeOpenListMountPath(value string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	if normalized == "" {
		return ""
	}

	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}

	normalized = strings.Join(strings.FieldsFunc(normalized, func(r rune) bool {
		return r == '/'
	}), "/")
	if normalized == "" {
		return "/"
	}

	return "/" + strings.Trim(strings.TrimSpace(normalized), "/")
}

// buildMountProviderByNetDisk 根据网络磁盘类型创建相应的Provider接口实现
func buildMountProviderByNetDisk(netDisk string, providerRepo *repository.ProviderRepository, mountRepo *repository.MountRepository) interfaces.Provider {
	switch netDisk {
	case "mock":
		return &providers.MockProvider{}
	case "baidu":
		return providers.NewBaiduProvider(providerRepo)
	case "quark":
		return providers.NewQuarkProvider(providerRepo)
	case "local":
		return providers.NewLocalProvider(providerRepo, mountRepo)
	default:
		return nil
	}
}
