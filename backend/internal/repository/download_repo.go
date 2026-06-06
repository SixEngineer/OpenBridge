package repository

import (
	"openbridge/backend/internal/domain/entity"

	"gorm.io/gorm"
)

type DownloadRepository struct {
	db *gorm.DB
}

func NewDownloadRepository(db *gorm.DB) *DownloadRepository {
	return &DownloadRepository{db: db}
}

func (repo *DownloadRepository) InsertTask(task *entity.DownloadTask) error {
	return repo.db.Create(task).Error
}

func (repo *DownloadRepository) GetTaskByTaskID(taskID string) (*entity.DownloadTask, error) {
	var task entity.DownloadTask
	if err := repo.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (repo *DownloadRepository) UpdateTask(task *entity.DownloadTask) error {
	return repo.db.Updates(task).Error
}
