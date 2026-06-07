package entity

import "time"

type MountPoint struct {
	ID                uint      `gorm:"column:id;primaryKey" json:"id"`
	Name              string    `gorm:"column:name;size:100;not null" json:"name"`
	ProviderAccountID uint      `gorm:"column:provider_account_id;not null;index" json:"provider_account_id"`
	ProviderType      string    `gorm:"column:provider_type;size:50;not null;index" json:"provider_type"`
	MountPath         string    `gorm:"column:mount_path;size:255;not null" json:"mount_path"`
	ProviderRootPath  string    `gorm:"column:provider_root_path;size:255;not null" json:"provider_root_path"`
	QuotaMode         string    `gorm:"column:quota_mode;size:20;not null;index" json:"quota_mode"`
	InheritFromID     *uint     `gorm:"column:inherit_from_id;index" json:"inherit_from_id"`
	VirtualTotal      int64     `gorm:"column:virtual_total" json:"virtual_total"`
	VirtualUsed       int64     `gorm:"column:virtual_used" json:"virtual_used"`
	ReadOnly          bool      `gorm:"column:read_only;not null;default:false" json:"read_only"`
	Enabled           bool      `gorm:"column:enabled;not null;default:true" json:"enabled"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}
