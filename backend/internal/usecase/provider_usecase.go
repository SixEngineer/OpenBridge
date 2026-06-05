package usecase

import (
	"errors"
	"openbridge/backend/internal/domain/entity"
	"openbridge/backend/internal/domain/providers"
	"openbridge/backend/internal/repository"
	"openbridge/backend/internal/tool"
)

type ProviderUseCase struct {
	ProviderRepo     *repository.ProviderRepository
	ProviderRegistry *tool.Registry
	MountRepo        *repository.MountRepository
}

// 构造函数
func NewProviderUseCase(providerRepo *repository.ProviderRepository, providerRegistry *tool.Registry, mountRepo *repository.MountRepository) *ProviderUseCase {
	return &ProviderUseCase{
		ProviderRepo:     providerRepo,
		ProviderRegistry: providerRegistry,
		MountRepo:        mountRepo,
	}
}

// 注册 Provider
func (p *ProviderUseCase) RegisterProvider(providerAccount entity.ProviderAccount) error {

	providerNetDisk := providerAccount.NetDisk

	switch providerNetDisk {
	case "mock":
		p.ProviderRegistry.Register(providerAccount.Name, &providers.MockProvider{})
	case "baidu":
		p.ProviderRegistry.Register(providerAccount.Name, providers.NewBaiduProvider(p.ProviderRepo))
	case "quark":
		p.ProviderRegistry.Register(providerAccount.Name, providers.NewQuarkProvider(p.ProviderRepo))
	case "local": // Windows本地存储Provider，在Linux环境下不编译
		p.ProviderRegistry.Register(providerAccount.Name, providers.NewLocalProvider(p.ProviderRepo, p.MountRepo))
	default:
		return errors.New("provider netdisk type undefined")
	}

	return p.ProviderRepo.InsertProviderAccount(&providerAccount)
}

// 删除 Provider
func (p *ProviderUseCase) DeleteProvider(id uint) error {

	// 首先获取ProviderAccount信息，以便从Registry中注销
	providerAccount, err := p.ProviderRepo.GetProviderAccount(id)
	if err != nil {
		return err
	}

	// 从Registry中注销Provider
	p.ProviderRegistry.Unregister(providerAccount.Name)

	return p.ProviderRepo.DeleteProviderAccount(id)
}

// 更新 Provider 信息
func (p *ProviderUseCase) UpdateProvider(providerAccount entity.ProviderAccount) error {

	providerAccountOld, err := p.ProviderRepo.GetProviderAccount(providerAccount.ID)
	if err != nil {
		return err
	}

	// 删除旧的Provider
	p.ProviderRegistry.Unregister(providerAccountOld.Name)

	// 注册新的Provider
	providerNetDisk := providerAccount.NetDisk

	switch providerNetDisk {
	case "mock":
		p.ProviderRegistry.Register(providerAccount.Name, &providers.MockProvider{})
	case "baidu":
		p.ProviderRegistry.Register(providerAccount.Name, providers.NewBaiduProvider(p.ProviderRepo))
	case "quark":
		p.ProviderRegistry.Register(providerAccount.Name, providers.NewQuarkProvider(p.ProviderRepo))
	case "local": // Windows本地存储Provider，在Linux环境下不编译
		p.ProviderRegistry.Register(providerAccount.Name, providers.NewLocalProvider(p.ProviderRepo, p.MountRepo))
	// case "local_windows": // Windows本地存储Provider，在Linux环境下不编译
	// 	p.ProviderRegistry.Register(providerAccount.Name, &providers.LocalWindowsProvider{})
	// case "local_linux": // Linux本地存储Provider，在Windows环境下不编译
	// 	p.ProviderRegistry.Register(providerAccount.Name, providers.NewLocalLinuxProvider(p.ProviderRepo))
	default:
		return errors.New("provider not found")
	}

	return p.ProviderRepo.UpdateProviderAccount(&providerAccount)
}

// 获取 Provider
func (p *ProviderUseCase) GetProvider(id uint) (*entity.ProviderAccount, error) {
	return p.ProviderRepo.GetProviderAccount(id)
}

// 获取 Provider 列表
func (p *ProviderUseCase) ListProvider() ([]entity.ProviderAccount, error) {
	return p.ProviderRepo.ListProviderAccounts()
}
