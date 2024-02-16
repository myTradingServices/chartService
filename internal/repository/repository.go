package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmfshirokan/chartService/internal/model"
)

// ADD validation:

type repository struct {
	dbpool *pgxpool.Pool
}

type Cnadle interface {
	Add(ctx context.Context, cand model.Candle) error
	Delete(ctx context.Context, symbol string) error
	Get(ctx context.Context, symbol, bidOrAsk string, from time.Time) ([]model.Candle, error)
}

func New(conn *pgxpool.Pool) Cnadle {
	return &repository{
		dbpool: conn,
	}
}

func (repo *repository) Add(ctx context.Context, candle model.Candle) error {
	_, err := repo.dbpool.Exec(
		ctx,
		"INSERT INTO trading.candles VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		candle.Symbol,
		candle.BidOrAsk,
		candle.Highest,
		candle.Lowest,
		candle.Open,
		candle.Close,
		candle.OpenTime,
		candle.CloseTime,
	)

	return err
}

func (repo *repository) Delete(ctx context.Context, symbol string) error {
	_, err := repo.dbpool.Exec(ctx, "DELETE FROM trading.candles WHERE symbol = $1", symbol)

	return err
}

func (repo *repository) Get(ctx context.Context, symbol, bidOrAsk string, from time.Time) ([]model.Candle, error) {
	rows, err := repo.dbpool.Query(ctx, "SELECT * FROM trading.candles WHERE symbol = $1 AND open_time >= $2 AND bid_or_ask = $3", symbol, from, bidOrAsk)
	if err != nil {
		return nil, err
	}

	candleArr := make([]model.Candle, 40)
	index := 0

	for rows.Next() {
		err = rows.Scan(candleArr[index])
		if err != nil {
			return nil, err
		}

		index++
	}

	if rows.Err() != nil {
		return candleArr, rows.Err()
	}

	return candleArr, nil
}
