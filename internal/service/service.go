package service

import (
	"context"
	"time"

	"github.com/mmfshirokan/chartService/internal/model"
	"github.com/mmfshirokan/chartService/internal/repository"
)

type service struct {
	repo repository.Interface
}

type Interface interface {
	Add(ctx context.Context, cand model.Candle) error
	Delete(ctx context.Context, symbol string, bidOrAsk model.PriceType) error
	Get(ctx context.Context, symbol string, interval time.Duration, bidOrAsk model.PriceType) ([]model.Candle, error)
}

func New(repo repository.Interface) Interface {
	return service{
		repo: repo,
	}
}

func (serv service) Add(ctx context.Context, cand model.Candle) error {
	return serv.Add(ctx, cand)
}

func (serv service) Delete(ctx context.Context, symbol string, bidOrAsk model.PriceType) error {
	return serv.Delete(ctx, symbol, bidOrAsk)
}

func (serv service) Get(ctx context.Context, symbol string, interval time.Duration, bidOrAsk model.PriceType) ([]model.Candle, error) {
	return serv.Get(ctx, symbol, interval, bidOrAsk)
}
