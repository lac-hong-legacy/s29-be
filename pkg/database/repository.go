package database

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, model interface{}) error
	FindByID(ctx context.Context, id uint, model interface{}) error
	FindAll(ctx context.Context, models interface{}) error
	Update(ctx context.Context, model interface{}) error
	Delete(ctx context.Context, model interface{}) error
	Where(ctx context.Context, query interface{}, args ...interface{}) *gorm.DB
}

type BaseRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &BaseRepository{db: db}
}

func (r *BaseRepository) Create(ctx context.Context, model interface{}) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *BaseRepository) FindByID(ctx context.Context, id uint, model interface{}) error {
	return r.db.WithContext(ctx).First(model, id).Error
}

func (r *BaseRepository) FindAll(ctx context.Context, models interface{}) error {
	return r.db.WithContext(ctx).Find(models).Error
}

func (r *BaseRepository) Update(ctx context.Context, model interface{}) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *BaseRepository) Delete(ctx context.Context, model interface{}) error {
	return r.db.WithContext(ctx).Delete(model).Error
}

func (r *BaseRepository) Where(ctx context.Context, query interface{}, args ...interface{}) *gorm.DB {
	return r.db.WithContext(ctx).Where(query, args...)
}
