package model

import (
	"time"

	// "github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

type Candle struct { // Main DTO.
	Symbol   string          `validate:"max=6"`       // Broker symbol.
	BidOrAsk PriceType       `validate:"oneof:0,1,2"` // Either bid or Ask.
	Highest  decimal.Decimal // Highest price value in given time interval.
	Lowest   decimal.Decimal // Lowest price value in given time interval.
	Open     decimal.Decimal // Price at the begining of time interval.
	Close    decimal.Decimal // Price at the end of time interval.
	OpenTime time.Time       // Time at the begining of time interval.
	Interval time.Duration   // Time Interval.
}

func (c Candle) GetCloseTime() time.Time {
	return c.OpenTime.Add(c.Interval)
}

// TODO: add validate method for Candle struct
