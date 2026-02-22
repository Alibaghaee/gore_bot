package repository

import (
	"context"
	"telebridge/internal/domain"
	"gorm.io/gorm"
)

type BotRepository struct {
	db *gorm.DB
}

func NewBotRepository(db *gorm.DB) *BotRepository {
	return &BotRepository{db: db}
}

func (r *BotRepository) Create(ctx context.Context, bot *domain.Bot) error {
	return r.db.WithContext(ctx).Create(bot).Error
}

func (r *BotRepository) GetBySlug(ctx context.Context, slug string) (*domain.Bot, error) {
	var bot domain.Bot
	err := r.db.WithContext(ctx).Where("slug = ? AND is_active = ?", slug, true).First(&bot).Error
	return &bot, err
}
