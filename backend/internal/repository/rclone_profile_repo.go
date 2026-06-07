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

func (repo *RcloneProfileRepository) Delete(id uint) error {
	return repo.db.Delete(&entity.RcloneProfile{}, id).Error
}

func (repo *RcloneProfileRepository) Get(id uint) (*entity.RcloneProfile, error) {
	var profile entity.RcloneProfile
	if err := repo.db.First(&profile, id).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (repo *RcloneProfileRepository) List() ([]entity.RcloneProfile, error) {
	var profiles []entity.RcloneProfile
	if err := repo.db.Order("updated_at desc").Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}
