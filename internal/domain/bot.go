package domain

import "time"

type Bot struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	Slug      string    `gorm:"uniqueIndex;not null" json:"slug"` // آدرس اختصاصی وب‌هوک
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
