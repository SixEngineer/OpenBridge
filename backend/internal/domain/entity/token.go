package entity

import "time"

// token实体类
type Token struct {
	ID           uint        `gorm:"column:id;primaryKey" json:"id"`
	NetDisk      string      `gorm:"column:net_disk;size:50;not null;index" json:"netDisk"`
	AccessToken  string      `gorm:"column:access_token;type:text;not null" json:"accessToken"`
	RefreshToken string      `gorm:"column:refresh_token;type:text" json:"refreshToken"`
	ExpiresAt    *time.Time  `gorm:"column:expires_at" json:"expiresAt"`
	CreatedAt    time.Time   `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time   `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

const (
	NetDiskBaidu   = "baidu"
	NetDiskAliyun  = "aliyun"
	NetDiskQuark   = "quark"
)