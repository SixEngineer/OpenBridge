package repository

import (
	"openbridge/backend/internal/domain/entity"

	"gorm.io/gorm"
)

// quota_snapshots表的Repository
type QuotaRepository struct {
	db *gorm.DB
}

// 构造 quota_snapshots 表的 Repo 对象
func NewQuotaRepository(db *gorm.DB) *QuotaRepository {
	return &QuotaRepository{db: db}
}

// 写入配额快照
func (repo *QuotaRepository) InsertQuotaSnapshot(snapshot *entity.QuotaSnapshot) error {
	return repo.db.Create(snapshot).Error
}
