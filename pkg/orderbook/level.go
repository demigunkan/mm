package orderbook

import "github.com/demigunkan/mm/internal/interfaces"

var _ interfaces.Level = &level{}

type level struct {
	price    float64
	amount   float64
	netPrice float64
}

func (l *level) Amount() float64 {
	return l.amount
}

func (l *level) Price() float64 {
	return l.price
}

func (l *level) NetPrice() float64 {
	return l.netPrice
}
