package model

type PriceType int

const (
	Ask PriceType = iota
	Bid PriceType = iota
)

func (p PriceType) String() string {
	return [...]string{"ask", "bid"}[p]
}
