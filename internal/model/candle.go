package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Candle struct {
	Symbol    string
	BidOrAsk  string
	Highest   decimal.Decimal
	Lowest    decimal.Decimal
	Open      decimal.Decimal
	Close     decimal.Decimal
	OpenTime  time.Time
	CloseTime time.Time
}
