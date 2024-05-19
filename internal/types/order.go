package types

type OrderType int

const (
	OrderType__MARKET OrderType = iota
	OrderType__LIMIT
)
