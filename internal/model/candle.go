package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Candle struct {
	Symbol   string    `validate:"max=6"`
	BidOrAsk PriceType `validate:"oneof:0,1"` // change to yota in go Type
	Highest  decimal.Decimal
	Lowest   decimal.Decimal
	Open     decimal.Decimal
	Close    decimal.Decimal
	OpenTime time.Time
	Interval time.Duration // change yo interval
}

func (c Candle) GetCloseTime() time.Time {
	return c.OpenTime.Add(c.Interval)
}
