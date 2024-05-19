package interfaces

import "github.com/demigunkan/mm/internal/types"

type Orderbook interface {
	Levels(types.Side) []Level
	Iterate(types.Side, func(Level) bool)
	ModifyLevel(types.Side, float64, float64, float64)
	Top(types.Side) Level
	TotalAmountAt(types.Side, float64) float64
	Print()
}
