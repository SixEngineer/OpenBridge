package entity

import "time"

type QuotaSnapshot struct {
	ID                uint      `gorm:"column:id;primaryKey" json:"id"`
	Provider          string    `gorm:"column:provider;size:100;not null;index" json:"provider"`
	MountPointID      uint      `gorm:"column:mount_point_id;index" json:"mount_point_id"`
	ProviderAccountID uint      `gorm:"column:provider_account_id;not null;index" json:"provider_account_id"`
	Mode              string    `gorm:"column:mode;size:20;not null;default:'real'" json:"mode"`
	Total             int64     `gorm:"column:total_quota" json:"total"`
	Used              int64     `gorm:"column:used_quota" json:"used"`
	Available         int64     `gorm:"column:available_quota" json:"available"`
	SyncStatus        string    `gorm:"column:sync_status;size:20;not null;default:'success'" json:"sync_status"`
	ErrorMessage      string    `gorm:"column:error_message;type:text" json:"error_message"`
	SyncedAt          time.Time `gorm:"column:synced_at;not null;index" json:"synced_at"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}
