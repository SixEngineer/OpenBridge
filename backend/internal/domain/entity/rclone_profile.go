package entity

import "time"

type RcloneProfile struct {
	ID              uint       `gorm:"column:id;primaryKey" json:"id"`
	OpenListBaseURL string     `gorm:"column:openlist_base_url;size:255;not null;default:'';uniqueIndex:idx_rclone_profile_scope_name;index" json:"openlist_base_url"`
	Name            string     `gorm:"column:name;size:120;not null;uniqueIndex:idx_rclone_profile_scope_name" json:"name"`
	Mode            string     `gorm:"column:mode;size:20;not null;index" json:"mode"`
	MountIDs        string     `gorm:"column:mount_ids;type:text;not null" json:"mount_ids"`
	ManagedRemotes  string     `gorm:"column:managed_remotes;type:text;not null;default:'[]'" json:"-"`
	Username        string     `gorm:"column:username;size:255;not null" json:"username"`
	PasswordCipher  string     `gorm:"column:password_cipher;type:text;not null" json:"-"`
	TargetPath      string     `gorm:"column:target_path;size:255;not null" json:"target_path"`
	IsMounted       bool       `gorm:"column:is_mounted;not null;default:false" json:"is_mounted"`
	MountPID        int        `gorm:"column:mount_pid;not null;default:0" json:"mount_pid"`
	MountRCAddr     string     `gorm:"column:mount_rc_addr;size:255" json:"mount_rc_addr"`
	LastAppliedAt   *time.Time `gorm:"column:last_applied_at" json:"last_applied_at"`
	LastMountedAt   *time.Time `gorm:"column:last_mounted_at" json:"last_mounted_at"`
	LastError       string     `gorm:"column:last_error;type:text" json:"last_error"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}
