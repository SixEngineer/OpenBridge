package entity

import "time"

type DownloadTask struct {
	ID              uint   `gorm:"primaryKey"`
	TaskID          string `gorm:"size:64;uniqueIndex;not null"`
	OpenListBaseURL string `gorm:"column:openlist_base_url;size:255;not null;default:'';index"`
	SourcePath      string `gorm:"type:text;not null"`
	// SourceType        string `gorm:"size:50"`
	FileName   string `gorm:"size:255"`
	FileSize   int64
	DirectLink string `gorm:"type:text"`
	// 下载完成后 aria2 返回的绝对路径，持久化到 DB 以免 aria2 遗忘后查不到
	FilePath string `gorm:"type:text"`
	// ProviderAccountID *uint   `gorm:"index"`
	// ProviderType      string  `gorm:"size:50;index"`
	Aria2GID        string  `gorm:"size:64;index"`
	Status          string  `gorm:"size:30;not null;index"`
	Progress        float64 `gorm:"default:0"`
	CompletedLength int64   `gorm:"-"`
	TotalLength     int64   `gorm:"-"`
	DownloadSpeed   int64   `gorm:"-"`
	ErrorMessage    string  `gorm:"type:text"`
	RetryCount      int     `gorm:"default:0"`
	StartedAt       *time.Time
	FinishedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
