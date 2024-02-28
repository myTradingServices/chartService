package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmfshirokan/chartService/internal/model"
	log "github.com/sirupsen/logrus"
)

// ADD validation:

type repository struct {
	dbpool *pgxpool.Pool
}

type Interface interface { // TODO mabe change (to )
	Add(ctx context.Context, cand model.Candle) error
	Delete(ctx context.Context, symbol string, bidOrAsk model.PriceType) error
	Get(ctx context.Context, symbol string, interval time.Duration, bidOrAsk model.PriceType) ([]model.Candle, error)
}

func New(conn *pgxpool.Pool) Interface {
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
		candle.Interval,
	)

	return err
}

func (repo *repository) Delete(ctx context.Context, symbol string, bidOrAsk model.PriceType) error {
	_, err := repo.dbpool.Exec(ctx, "DELETE FROM trading.candles WHERE symbol = $1 AND bid_or_ask = $2", symbol, bidOrAsk)

	return err
}

func (repo *repository) Get(ctx context.Context, symbol string, interval time.Duration, bidOrAsk model.PriceType) ([]model.Candle, error) {
	rows, err := repo.dbpool.Query(ctx, "SELECT * FROM trading.candles WHERE symbol = $1 AND time_interval = $2 AND bid_or_ask = $3", symbol, interval, bidOrAsk)
	if err != nil {
		return nil, err
	}

	candleArr := []model.Candle{}
	index := 0

	for rows.Next() {
		tmpCandle := model.Candle{}

		tmpString := ""

		err = rows.Scan(
			&tmpCandle.Symbol,
			&tmpString, //&candleArr[index].BidOrAsk,
			&tmpCandle.Highest,
			&tmpCandle.Lowest,
			&tmpCandle.Open,
			&tmpCandle.Close,
			&tmpCandle.OpenTime,
			&tmpCandle.Interval,
		)
		if err != nil {
			return nil, err
		}

		AddPriceType(tmpString, &tmpCandle)

		candleArr = append(candleArr, tmpCandle)

		index++
	}

	if rows.Err() != nil {
		return candleArr, rows.Err()
	}

	return candleArr, nil
}

func AddPriceType(str string, c *model.Candle) {
	if str == "Ask" || str == "ask" {
		c.BidOrAsk = model.Ask
		return
	}

	if str == "Bid" || str == "bid" {
		c.BidOrAsk = model.Bid
		return
	}

	log.Error("Neither Bid or Ask is set.")
}
