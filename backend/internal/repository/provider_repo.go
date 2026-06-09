package repository

import (
	"openbridge/backend/internal/domain/entity"
	"time"

	"gorm.io/gorm"
)

type ProviderRepository struct {
	db *gorm.DB
}

// 构造函数
func NewProviderRepository(db *gorm.DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

// 创建 ProviderAccount
func (repo *ProviderRepository) InsertProviderAccount(providerAccount *entity.ProviderAccount) error {
	return repo.db.Create(providerAccount).Error
}

// 删除 ProviderAccount
func (repo *ProviderRepository) DeleteProviderAccount(id uint) error {
	return repo.db.Delete(&entity.ProviderAccount{}, id).Error
}

// 更新 ProviderAccount
func (repo *ProviderRepository) UpdateProviderAccount(providerAccount *entity.ProviderAccount) error {
	if providerAccount == nil || providerAccount.ID == 0 {
		return gorm.ErrMissingWhereClause
	}

	return repo.db.Model(&entity.ProviderAccount{}).
		Where("id = ?", providerAccount.ID).
		Select(
			"name",
			"openlist_base_url",
			"provider_type",
			"net_disk",
			"account_id",
			"status",
			"access_token",
			"refresh_token",
			"token_expires_at",
			"auth_cookie",
			"total_quota",
			"used_quota",
			"available_quota",
			"last_quota_sync_at",
			"last_error",
		).
		Updates(providerAccount).Error
}

// 获取 ProviderAccount
func (repo *ProviderRepository) GetProviderAccount(id uint) (*entity.ProviderAccount, error) {
	var providerAccount entity.ProviderAccount
	err := repo.db.First(&providerAccount, id).Error
	if err != nil {
		return nil, err
	}
	return &providerAccount, nil
}

// 获取所有 ProviderAccount
func (repo *ProviderRepository) ListProviderAccounts() ([]entity.ProviderAccount, error) {
	var providerAccounts []entity.ProviderAccount
	err := repo.db.Find(&providerAccounts).Error
	if err != nil {
		return nil, err
	}
	return providerAccounts, nil
}

func (repo *ProviderRepository) AssignEmptyOpenListScope(scope string) error {
	return repo.db.Model(&entity.ProviderAccount{}).
		Where("openlist_base_url = '' OR openlist_base_url IS NULL").
		Update("openlist_base_url", scope).Error
}

func (repo *ProviderRepository) GetProviderAccountByOpenList(id uint, scope string) (*entity.ProviderAccount, error) {
	var providerAccount entity.ProviderAccount
	err := repo.db.Where("id = ? AND openlist_base_url = ?", id, scope).First(&providerAccount).Error
	if err != nil {
		return nil, err
	}
	return &providerAccount, nil
}

func (repo *ProviderRepository) ListProviderAccountsByOpenList(scope string) ([]entity.ProviderAccount, error) {
	var providerAccounts []entity.ProviderAccount
	err := repo.db.Where("openlist_base_url = ?", scope).Find(&providerAccounts).Error
	if err != nil {
		return nil, err
	}
	return providerAccounts, nil
}

// 按 provider 标识查询 ProviderAccount，优先按 name，兜底按 net_disk
func (repo *ProviderRepository) GetProviderAccountByProvider(provider string) (*entity.ProviderAccount, error) {
	var providerAccount entity.ProviderAccount
	err := repo.db.Where("name = ?", provider).First(&providerAccount).Error
	if err == nil {
		return &providerAccount, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	err = repo.db.Where("net_disk = ?", provider).First(&providerAccount).Error
	if err != nil {
		return nil, err
	}

	return &providerAccount, nil
}

// 更新 ProviderAccount 当前配额
func (repo *ProviderRepository) UpdateProviderQuota(providerAccountID uint, total int64, used int64, available int64, syncedAt time.Time) error {
	updates := map[string]interface{}{
		"total_quota":        total,
		"used_quota":         used,
		"available_quota":    available,
		"last_quota_sync_at": syncedAt,
		"updated_at":         syncedAt,
	}

	return repo.db.Model(&entity.ProviderAccount{}).Where("id = ?", providerAccountID).Updates(updates).Error
}
