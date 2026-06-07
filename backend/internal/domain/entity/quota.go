package entity

import "time"

type Quota struct {
	Provider  string    `json:"provider"`
	Total     int64     `json:"total"`
	Used      int64     `json:"used"`
	Available int64     `json:"available"`
	UpdatedAt time.Time `json:"updated_at"`
}