package storage

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

type store struct {
	DB *gorm.DB
}

func NewStore(db *gorm.DB) *store {
	return &store{
		DB: db,
	}
}

type Stock struct {
	gorm.Model
	Quantity int32 `gorm:"not null"`
	PriceId  string
}

func (s *store) GetItems(ctx context.Context, ids []int32) ([]*Stock, error) {
	var stocks []*Stock
	err := s.DB.Where("id IN (?)", ids).Find(&stocks).Error
	if err != nil {
		return nil, err
	}
	return stocks, nil
}
func (s *store) UpdateStock(ctx context.Context, stocks []*Stock) error {
	if stocks == nil || len(stocks) == 0 {
		return errors.New("no stock")
	}
	err := s.DB.Save(stocks).Error
	if err != nil {
		return err
	}
	return nil
}
