package entity

import "time"

type RcloneProfile struct {
	ID             uint       `gorm:"column:id;primaryKey" json:"id"`
	Name           string     `gorm:"column:name;size:120;not null;uniqueIndex" json:"name"`
	Mode           string     `gorm:"column:mode;size:20;not null;index" json:"mode"`
	MountIDs       string     `gorm:"column:mount_ids;type:text;not null" json:"mount_ids"`
	Username       string     `gorm:"column:username;size:255;not null" json:"username"`
	PasswordCipher string     `gorm:"column:password_cipher;type:text;not null" json:"-"`
	TargetPath     string     `gorm:"column:target_path;size:255;not null" json:"target_path"`
	LastAppliedAt  *time.Time `gorm:"column:last_applied_at" json:"last_applied_at"`
	LastMountedAt  *time.Time `gorm:"column:last_mounted_at" json:"last_mounted_at"`
	LastError      string     `gorm:"column:last_error;type:text" json:"last_error"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}
