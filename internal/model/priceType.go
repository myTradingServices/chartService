package model

type PriceType int

const (
	Unknown PriceType = iota
	Ask     PriceType = iota
	Bid     PriceType = iota
)
