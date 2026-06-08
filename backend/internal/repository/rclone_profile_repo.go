package repository

import (
	"openbridge/backend/internal/domain/entity"

	"gorm.io/gorm"
)

type RcloneProfileRepository struct {
	db *gorm.DB
}

func NewRcloneProfileRepository(db *gorm.DB) *RcloneProfileRepository {
	return &RcloneProfileRepository{db: db}
}

func (repo *RcloneProfileRepository) Create(profile *entity.RcloneProfile) error {
	return repo.db.Create(profile).Error
}

func (repo *RcloneProfileRepository) Update(profile *entity.RcloneProfile) error {
	return repo.db.Save(profile).Error
}

func (repo *RcloneProfileRepository) Delete(id uint, scope string) error {
	return repo.db.Where("id = ? AND openlist_base_url = ?", id, scope).Delete(&entity.RcloneProfile{}).Error
}

func (repo *RcloneProfileRepository) Get(id uint, scope string) (*entity.RcloneProfile, error) {
	var profile entity.RcloneProfile
	if err := repo.db.Where("id = ? AND openlist_base_url = ?", id, scope).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (repo *RcloneProfileRepository) List(scope string) ([]entity.RcloneProfile, error) {
	var profiles []entity.RcloneProfile
	if err := repo.db.Where("openlist_base_url = ?", scope).Order("updated_at desc").Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

func (repo *RcloneProfileRepository) AssignEmptyOpenListScope(scope string) error {
	return repo.db.
		Model(&entity.RcloneProfile{}).
		Where("openlist_base_url = '' OR openlist_base_url IS NULL").
		Update("openlist_base_url", scope).Error
}
