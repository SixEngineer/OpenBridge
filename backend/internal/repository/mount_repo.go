package repository

import (
	"openbridge/backend/internal/domain/entity"

	"gorm.io/gorm"
)

type MountRepository struct {
	db *gorm.DB
}

func NewMountRepository(db *gorm.DB) *MountRepository {
	return &MountRepository{db: db}
}

func (repo *MountRepository) InsertMountPoint(mountPoint *entity.MountPoint) error {
	return repo.db.Create(mountPoint).Error
}

func (repo *MountRepository) GetMountPoint(id uint) (*entity.MountPoint, error) {
	var mount entity.MountPoint
	if err := repo.db.First(&mount, id).Error; err != nil {
		return nil, err
	}
	return &mount, nil
}

func (repo *MountRepository) GetMountPointByProviderAccountID(providerAccountID uint) (*entity.MountPoint, error) {
	var mount entity.MountPoint
	if err := repo.db.Where("provider_account_id = ?", providerAccountID).First(&mount).Error; err != nil {
		return nil, err
	}
	return &mount, nil
}

func (repo *MountRepository) ListMountPointsByProviderAccountID(providerAccountID uint) ([]entity.MountPoint, error) {
	var mounts []entity.MountPoint
	if err := repo.db.Where("provider_account_id = ?", providerAccountID).Find(&mounts).Error; err != nil {
		return nil, err
	}
	return mounts, nil
}

func (repo *MountRepository) ListAllMountPoints() ([]entity.MountPoint, error) {
	var mounts []entity.MountPoint
	if err := repo.db.Find(&mounts).Error; err != nil {
		return nil, err
	}
	return mounts, nil
}

func (repo *MountRepository) UpdateMountPoint(mount *entity.MountPoint) error {
	return repo.db.Save(mount).Error
}

func (repo *MountRepository) DeleteMountPoint(id uint) error {
	return repo.db.Delete(&entity.MountPoint{}, id).Error
}

func (repo *MountRepository) DeleteMountPointsByProviderAccountID(providerAccountID uint) error {
	return repo.db.Where("provider_account_id = ?", providerAccountID).Delete(&entity.MountPoint{}).Error
}

// DeleteInheritMountPointsByParentProvider 删除继承自指定 provider 下所有挂载点的继承挂载点
func (repo *MountRepository) DeleteInheritMountPointsByParentProvider(providerAccountID uint) error {
	subQuery := repo.db.Model(&entity.MountPoint{}).Select("id").Where("provider_account_id = ?", providerAccountID)
	return repo.db.Where("inherit_from_id IN (?)", subQuery).Delete(&entity.MountPoint{}).Error
}
