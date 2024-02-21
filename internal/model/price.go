package model

import (
	"github.com/golang/protobuf/ptypes/timestamp"
)

type Price struct {
	Bid    Decim
	Ask    Decim
	Symbol string
	Date   *timestamp.Timestamp
}

type Decim struct {
	Value int64
	Exp   int32
}
