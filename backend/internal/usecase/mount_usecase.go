package usecase

import (
	"context"
	"errors"
	"fmt"
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
	ErrMountProviderRequired      = errors.New("real mode requires provider_account_id")
	ErrMountParentRequired        = errors.New("inherit mode requires inherit_from_id")
	ErrMountParentNotReal         = errors.New("inherit parent must be real mode")
	ErrMountCircularInherit       = errors.New("inherit chain has cycle")
	ErrMountVirtualExceedsAllowed = errors.New("virtual_total exceeds allowed_max")
	ErrMountVirtualUsedInvalid    = errors.New("virtual_used must be <= virtual_total")
	ErrMountDisabled              = errors.New("mount is disabled")
	ErrProviderNotFound           = errors.New("provider not found")
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
}

func NewMountUseCase(mountRepo *repository.MountRepository, providerRepo *repository.ProviderRepository, quotaRepo *repository.QuotaRepository, providerRegistry *tool.Registry) *MountUseCase {
	return &MountUseCase{
		mountRepo:        mountRepo,
		providerRepo:     providerRepo,
		quotaRepo:        quotaRepo,
		providerRegistry: providerRegistry,
	}
}

// CreateMount 创建一个新的挂载点
// ctx: 上下文信息，用于传递请求范围的数据和控制信号
// mount: 包含挂载点配置信息的实体
// 返回值: 成功时返回创建的挂载点指针，失败时返回错误信息
func (u *MountUseCase) CreateMount(ctx context.Context, mount entity.MountPoint) (*entity.MountPoint, error) {
	// 将配额模式转换为统一的小写格式并去除前后空格
	mode := entity.QuotaMode(strings.ToLower(strings.TrimSpace(mount.QuotaMode)))
	// 更新挂载点的配额模式为处理后的格式
	mount.QuotaMode = string(mode)

	// 验证挂载点配置是否有效
	if err := u.validateMountConfig(ctx, &mount, mode); err != nil {
		return nil, err
	}

	// 如果是虚拟配额模式且虚拟总量为0，则设置为只读模式
	if mode == entity.QuotaModeVirtual && mount.VirtualTotal == 0 {
		mount.ReadOnly = true
	}

	// 将挂载点信息插入到存储库中
	if err := u.mountRepo.InsertMountPoint(&mount); err != nil {
		return nil, err
	}
	// 返回创建成功的挂载点信息
	return &mount, nil
}

func (u *MountUseCase) GetMountQuota(ctx context.Context, mountID uint) (MountQuotaResult, error) {
	return u.resolveMountQuota(ctx, mountID, false)
}

func (u *MountUseCase) SyncMountQuota(ctx context.Context, mountID uint) (MountQuotaResult, error) {
	return u.resolveMountQuota(ctx, mountID, true)
}

// resolveMountQuota 是一个用于解析挂载点配额的方法
// 它接收上下文、挂载点ID和是否同步远程的标志作为参数
// 返回解析结果和可能的错误
func (u *MountUseCase) resolveMountQuota(ctx context.Context, mountID uint, syncRemote bool) (MountQuotaResult, error) {
	// 从仓库获取挂载点信息
	mount, err := u.mountRepo.GetMountPoint(mountID)
	if err != nil {
		return MountQuotaResult{}, err
	}
	// 检查挂载点是否启用，未启用则返回错误
	if !mount.Enabled {
		return MountQuotaResult{}, ErrMountDisabled
	}

	// 根据模式解析配额，使用map防止循环引用
	result, err := u.resolveByMode(ctx, mount, syncRemote, map[uint]struct{}{})
	if err != nil {
		// 记录解析失败的日志
		logger.L().Error("mount quota resolve failed",
			zap.Uint("mount_id", mount.ID),
			zap.String("mode", mount.QuotaMode),
			zap.Error(err),
		)
		now := time.Now().UTC()
		// 插入配额快照记录失败状态
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
	// 插入成功的配额快照记录
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

	// 更新配额的更新时间并记录成功日志
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

// resolveByMode 根据挂载点的配额模式递归解析最终配额结果。
// 对于 real 模式直接读取真实配额；inherit 模式沿父挂载点继承；virtual 模式按虚拟配置计算。
// visited 用于检测继承链循环，避免递归陷入死循环。
func (u *MountUseCase) resolveByMode(ctx context.Context, mount *entity.MountPoint, syncRemote bool, visited map[uint]struct{}) (MountQuotaResult, error) {
	// 进入递归前先检查当前节点是否已访问，已访问说明继承链出现环。
	if _, ok := visited[mount.ID]; ok {
		return MountQuotaResult{}, ErrMountCircularInherit
	}
	// 标记当前节点为已访问，供后续递归分支进行环检测。
	visited[mount.ID] = struct{}{}

	// 将配额模式统一转为小写，避免大小写导致分支判断不一致。
	mode := entity.QuotaMode(strings.ToLower(mount.QuotaMode))
	switch mode {
	case entity.QuotaModeReal:
		// real：直接解析真实账户配额（可选是否同步远端）。
		quota, err := u.resolveRealQuota(ctx, mount, syncRemote)
		if err != nil {
			return MountQuotaResult{}, err
		}
		// 输出命中分支日志，便于排查配额来源。
		logger.L().Info("quota resolver mode",
			zap.String("mode", string(mode)),
			zap.Uint("mount_id", mount.ID),
		)
		// real 模式下，AllowedMax 取真实 total。
		return MountQuotaResult{
			MountID:    mount.ID,
			Mode:       mount.QuotaMode,
			AllowedMax: quota.Total,
			Quota:      quota,
		}, nil
	case entity.QuotaModeInherit:
		// inherit：必须配置父挂载点，否则无法继承。
		if mount.InheritFromID == nil {
			return MountQuotaResult{}, ErrMountParentRequired
		}
		// 拉取父挂载点信息。
		parent, err := u.mountRepo.GetMountPoint(*mount.InheritFromID)
		if err != nil {
			return MountQuotaResult{}, err
		}
		// 当前规则仅允许继承 real 模式父节点，避免多级策略复杂化。
		if strings.ToLower(parent.QuotaMode) != string(entity.QuotaModeReal) {
			return MountQuotaResult{}, ErrMountParentNotReal
		}
		// 递归解析父节点配额；visited 在递归链路中共享，用于统一环检测。
		parentResult, err := u.resolveByMode(ctx, parent, syncRemote, visited)
		if err != nil {
			return MountQuotaResult{}, err
		}
		// 当前仅返回直接父节点 ID 作为继承链信息。
		chain := []uint{parent.ID}
		// 记录继承路径日志，方便追踪解析链路。
		logger.L().Info("quota resolver inherit chain",
			zap.Uint("mount_id", mount.ID),
			zap.Uint("parent_mount_id", parent.ID),
		)
		// inherit 直接复用父节点的 AllowedMax 与 Quota。
		return MountQuotaResult{
			MountID:      mount.ID,
			Mode:         mount.QuotaMode,
			AllowedMax:   parentResult.AllowedMax,
			Quota:        parentResult.Quota,
			InheritChain: chain,
		}, nil
	case entity.QuotaModeVirtual:
		// virtual：先计算允许上限（通常来源于真实账户 total）。
		allowedMax, err := u.getAllowedMax(ctx, mount, syncRemote)
		if err != nil {
			return MountQuotaResult{}, err
		}
		// 基础约束：虚拟已用量不能大于虚拟总量。
		if mount.VirtualUsed > mount.VirtualTotal {
			return MountQuotaResult{}, ErrMountVirtualUsedInvalid
		}
		// 安全约束：虚拟总量不能突破允许上限。
		if mount.VirtualTotal > allowedMax {
			logger.L().Warn("virtual quota validation failed",
				zap.Uint("mount_id", mount.ID),
				zap.Int64("virtual_total", mount.VirtualTotal),
				zap.Int64("allowed_max", allowedMax),
			)
			return MountQuotaResult{}, ErrMountVirtualExceedsAllowed
		}
		// 记录 virtual 分支计算参数，便于线上审计与调试。
		logger.L().Info("quota resolver mode",
			zap.String("mode", string(mode)),
			zap.Uint("mount_id", mount.ID),
			zap.Int64("virtual_total", mount.VirtualTotal),
			zap.Int64("virtual_used", mount.VirtualUsed),
			zap.Int64("allowed_max", allowedMax),
		)
		// 由虚拟配置直接构造返回配额，不读取真实账户已用值。
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
			// 回传虚拟配置明细，便于调用方展示和二次校验。
			VirtualConfig: map[string]int64{
				"virtual_total": mount.VirtualTotal,
				"virtual_used":  mount.VirtualUsed,
			},
		}, nil
	default:
		// 未识别模式统一返回模式非法错误。
		return MountQuotaResult{}, ErrMountInvalidMode
	}
}

// resolveRealQuota 用于解析和获取真实的配额信息
// ctx: 上下文信息
// mount: 挂载点实体
// syncRemote: 是否同步远程配额
// 返回: 解析后的配额信息和可能的错误
func (u *MountUseCase) resolveRealQuota(ctx context.Context, mount *entity.MountPoint, syncRemote bool) (entity.Quota, error) {
	// 检查是否设置了提供商账户ID，如果没有则返回错误
	if mount.ProviderAccountID == 0 {
		return entity.Quota{}, ErrMountProviderRequired
	}
	// 从提供商仓库获取提供商账户信息
	account, err := u.providerRepo.GetProviderAccount(mount.ProviderAccountID)
	if err != nil {
		return entity.Quota{}, err
	}

	// 如果不需要同步远程配额
	if !syncRemote {
		// 验证存储的配额数据是否有效
		if account.TotalQuota < account.UsedQuota || account.TotalQuota < 0 || account.UsedQuota < 0 {
			return entity.Quota{}, fmt.Errorf("invalid stored quota in provider account")
		}
		// 返回本地存储的配额信息
		return entity.Quota{
			Provider:  mount.ProviderType,
			Total:     account.TotalQuota,
			Used:      account.UsedQuota,
			Available: account.AvailableQuota,
			UpdatedAt: account.UpdatedAt.UTC(),
		}, nil
	}

	// 如果需要同步远程配额
	// 解析提供商实例
	providerInstance, err := u.resolveProvider(account)
	if err != nil {
		return entity.Quota{}, err
	}
	// 从远程获取配额信息
	remoteQuota, err := providerInstance.GetQuota(ctx, account)
	if err != nil {
		return entity.Quota{}, err
	}
	// 记录获取远程配额的日志
	logger.L().Info("provider quota fetched",
		zap.String("provider", account.NetDisk),
		zap.Uint("provider_account_id", account.ID),
		zap.Int64("total", remoteQuota.Total),
		zap.Int64("used", remoteQuota.Used),
		zap.Int64("available", remoteQuota.Available),
	)

	// 获取当前UTC时间
	now := time.Now().UTC()
	// 更新提供商配额信息到本地存储
	if err := u.providerRepo.UpdateProviderQuota(account.ID, remoteQuota.Total, remoteQuota.Used, remoteQuota.Available, now); err != nil {
		return entity.Quota{}, err
	}
	// 设置提供商类型和更新时间后返回远程配额信息
	remoteQuota.Provider = mount.ProviderType
	remoteQuota.UpdatedAt = now
	return remoteQuota, nil
}

// getAllowedMax 是一个方法，用于获取允许的最大挂载配额
// 它接收上下文、挂载点指针和是否同步远程的布尔值作为参数
// 返回允许的最大配额和可能的错误
func (u *MountUseCase) getAllowedMax(ctx context.Context, mount *entity.MountPoint, syncRemote bool) (int64, error) {
	// 从提供者仓库获取提供者账户信息
	account, err := u.providerRepo.GetProviderAccount(mount.ProviderAccountID)
	if err != nil {
		return 0, err
	}
	// 如果不需要同步远程且账户总配额大于0，则直接返回账户总配额
	if !syncRemote && account.TotalQuota > 0 {
		return account.TotalQuota, nil
	}

	// 解析提供者实例
	providerInstance, err := u.resolveProvider(account)
	if err != nil {
		return 0, err
	}
	// 从提供者实例获取配额信息
	quota, err := providerInstance.GetQuota(ctx, account)
	if err != nil {
		return 0, err
	}
	// 获取当前UTC时间
	now := time.Now().UTC()
	// 更新提供者配额信息到数据库
	if err := u.providerRepo.UpdateProviderQuota(account.ID, quota.Total, quota.Used, quota.Available, now); err != nil {
		return 0, err
	}
	// 记录虚拟允许最大配额评估日志
	logger.L().Info("virtual allowed_max evaluated",
		zap.Uint("mount_id", mount.ID),
		zap.Int64("allowed_max", quota.Total),
	)
	// 返回配额总量
	return quota.Total, nil
}

// validateMountConfig 是一个验证挂载配置的方法，根据不同的配额模式进行相应的验证
// ctx: 上下文信息，用于传递请求范围的数据和控制信号
// mount: 挂载点实体，包含挂载相关的配置信息
// mode: 配额模式，包括实时模式、继承模式和虚拟模式
// 返回值：error，验证过程中出现的错误
func (u *MountUseCase) validateMountConfig(ctx context.Context, mount *entity.MountPoint, mode entity.QuotaMode) error {
	// 根据不同的配额模式进行验证和处理
	switch mode {
	case entity.QuotaModeReal:
		// 实时模式验证：需要提供提供商账户ID
		if mount.ProviderAccountID == 0 {
			return ErrMountProviderRequired
		}
		// 获取提供商账户信息
		account, err := u.providerRepo.GetProviderAccount(mount.ProviderAccountID)
		if err != nil {
			return err
		}
		// 设置提供商类型，继承ID和虚拟总量、使用量为0
		mount.ProviderType = account.NetDisk
		mount.InheritFromID = nil
		mount.VirtualTotal = 0
		mount.VirtualUsed = 0
		return nil
	case entity.QuotaModeInherit:
		// 继承模式验证：需要提供父挂载点ID
		if mount.InheritFromID == nil {
			return ErrMountParentRequired
		}
		// 获取父挂载点信息
		parent, err := u.mountRepo.GetMountPoint(*mount.InheritFromID)
		if err != nil {
			return err
		}
		// 验证父挂载点必须是实时模式
		if strings.ToLower(parent.QuotaMode) != string(entity.QuotaModeReal) {
			return ErrMountParentNotReal
		}
		// 验证继承关系不会形成循环
		if err := u.validateNoCycle(parent.ID, mount.ID); err != nil {
			return err
		}
		// 设置提供商账户ID、类型和虚拟总量、使用量为0
		mount.ProviderAccountID = parent.ProviderAccountID
		mount.ProviderType = parent.ProviderType
		mount.VirtualTotal = 0
		mount.VirtualUsed = 0
		return nil
	case entity.QuotaModeVirtual:
		// 虚拟模式验证：需要提供提供商账户ID
		if mount.ProviderAccountID == 0 {
			return ErrMountProviderRequired
		}
		// 验证虚拟使用量不能超过虚拟总量
		if mount.VirtualUsed > mount.VirtualTotal {
			return ErrMountVirtualUsedInvalid
		}
		// 获取提供商账户信息
		account, err := u.providerRepo.GetProviderAccount(mount.ProviderAccountID)
		if err != nil {
			return err
		}
		// 设置提供商类型
		mount.ProviderType = account.NetDisk
		// 获取允许的最大虚拟配额并验证
		allowedMax, err := u.getAllowedMax(ctx, mount, true)
		if err != nil {
			return err
		}
		if mount.VirtualTotal > allowedMax {
			return ErrMountVirtualExceedsAllowed
		}
		return nil
	default:
		// 无效的配额模式
		return ErrMountInvalidMode
	}
}

// validateNoCycle 检查挂载点是否存在继承循环
// @param startID 起始挂载点ID
// @param candidateID 候选挂载点ID
// @return error 如果存在循环则返回ErrMountCircularInherit错误，否则返回nil
func (u *MountUseCase) validateNoCycle(startID uint, candidateID uint) error {
	// visited 用于记录已经访问过的挂载点ID，防止重复访问
	visited := map[uint]struct{}{}
	currentID := startID
	// 遍历挂载点继承链，直到遇到根挂载点(currentID为0)或出现错误
	for currentID != 0 {
		// 如果当前ID等于候选ID，说明存在循环继承
		if currentID == candidateID {
			return ErrMountCircularInherit
		}
		// 如果当前ID已经被访问过，说明存在循环继承
		if _, ok := visited[currentID]; ok {
			return ErrMountCircularInherit
		}
		// 将当前ID标记为已访问
		visited[currentID] = struct{}{}
		// 获取当前挂载点信息
		current, err := u.mountRepo.GetMountPoint(currentID)
		if err != nil {
			// 如果记录不存在，说明继承链已中断，返回nil
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			// 其他错误直接返回
			return err
		}
		// 如果没有继承自其他挂载点，说明到达继承链末端，返回nil
		if current.InheritFromID == nil {
			return nil
		}
		// 继续检查继承链上的下一个挂载点
		currentID = *current.InheritFromID
	}
	return nil
}

func (u *MountUseCase) resolveProvider(account *entity.ProviderAccount) (interfaces.Provider, error) {
	if providerInstance, ok := u.providerRegistry.Get(account.Name); ok {
		return providerInstance, nil
	}

	providerInstance := buildMountProviderByNetDisk(account.NetDisk, u.providerRepo, u.mountRepo)
	if providerInstance == nil {
		return nil, ErrProviderNotFound
	}
	_ = u.providerRegistry.Register(account.Name, providerInstance)
	return providerInstance, nil
}

// buildMountProviderByNetDisk 根据网络磁盘类型创建相应的Provider接口实现
// 参数:
//	netDisk: 网络磁盘类型字符串，如"mock"、"baidu"等
// 返回值:
//	interfaces.Provider: 根据输入返回对应的Provider接口实现，如果类型不支持则返回nil
func buildMountProviderByNetDisk(netDisk string, providerRepo *repository.ProviderRepository, mountRepo *repository.MountRepository) interfaces.Provider {
	switch netDisk {
	case "mock": // 如果是mock类型，返回MockProvider的实例
		return &providers.MockProvider{}
	case "baidu": // 如果是baidu类型，返回BaiduProvider的新实例
		return providers.NewBaiduProvider(providerRepo)
	case "quark": // 如果是quark类型，返回QuarkProvider的新实例
		return providers.NewQuarkProvider(providerRepo)
	case "local": // Windows本地存储Provider，在Linux环境下不编译
		return providers.NewLocalProvider(providerRepo, mountRepo)
	default: // 其他不支持的类型返回nil
		return nil
	}
}
