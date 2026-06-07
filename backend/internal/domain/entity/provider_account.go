package entity

import "time"

type ProviderAccount struct {
	ID              uint       `gorm:"column:id;primaryKey" json:"id"`
	Name            string     `gorm:"column:name;size:100;not null" json:"name"`
	OpenListBaseURL string     `gorm:"column:openlist_base_url;size:255;not null;default:'';index" json:"openlist_base_url"`
	ProviderType    string     `gorm:"column:provider_type;size:50;not null;index" json:"provider_type"`
	NetDisk         string     `gorm:"column:net_disk;size:50;not null" json:"net_disk"`
	AccountID       string     `gorm:"column:account_id;size:100" json:"account_id"`
	Status          string     `gorm:"column:status;size:20;not null;default:'active'" json:"status"`
	AccessToken     string     `gorm:"column:access_token;type:text" json:"access_token"`
	RefreshToken    string     `gorm:"column:refresh_token;type:text" json:"refresh_token"`
	TokenExpiresAt  *time.Time `gorm:"column:token_expires_at" json:"token_expires_at"`
	AuthCookie      string     `gorm:"column:auth_cookie;type:text" json:"auth_cookie"`
	TotalQuota      int64      `gorm:"column:total_quota" json:"total_quota"`
	UsedQuota       int64      `gorm:"column:used_quota" json:"used_quota"`
	AvailableQuota  int64      `gorm:"column:available_quota" json:"available_quota"`
	LastQuotaSyncAt *time.Time `gorm:"column:last_quota_sync_at" json:"last_quota_sync_at"`
	LastError       string     `gorm:"column:last_error;type:text" json:"last_error"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}
