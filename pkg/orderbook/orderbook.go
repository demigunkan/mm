package orderbook

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/demigunkan/mm/internal/interfaces"
	"github.com/demigunkan/mm/internal/types"
	"github.com/google/btree"
)

var _ interfaces.Orderbook = &Orderbook{}

type Orderbook struct {
	sides map[types.Side]*side
}

func New() interfaces.Orderbook {
	orderbook := &Orderbook{
		sides: make(map[types.Side]*side),
	}
	orderbook.sides[types.Side__ASK] = newSide(types.Side__ASK)
	orderbook.sides[types.Side__BID] = newSide(types.Side__BID)

	return orderbook
}

func (o *Orderbook) Levels(side types.Side) []interfaces.Level {
	var levels []interfaces.Level
	o.sides[side].tree.Descend(func(item interfaces.Level) bool {
		levels = append(levels, item)
		return true
	})

	return levels
}

func (o *Orderbook) Depth(side types.Side) int {
	return o.sides[side].tree.Len()
}

func (o *Orderbook) Top(side types.Side) interfaces.Level {
	if l, ok := o.sides[side].tree.Max(); ok {
		return l
	}
	return nil
}

func (o *Orderbook) Iterate(side types.Side, fn func(item interfaces.Level) bool) {
	if side == types.Side__BID {
		o.sides[side].tree.Descend(fn)
	} else {
		o.sides[side].tree.Ascend(fn)
	}
}

func (o *Orderbook) ModifyLevel(side types.Side, price float64, netPrice float64, amount float64) {
	priceKey := strconv.FormatFloat(price, 'f', 6, 64)

	if l, ok := o.sides[side].levels[priceKey]; !ok {
		if amount == 0 {
			return
		}
		l := &level{
			price:    price,
			netPrice: netPrice,
			amount:   amount,
		}
		o.sides[side].tree.ReplaceOrInsert(l)
		o.sides[side].levels[priceKey] = l
		return
	} else {
		if amount == 0 {
			o.sides[side].tree.Delete(l)
			delete(o.sides[side].levels, priceKey)
			return
		}

		l.amount = amount
		return
	}
}

func (o *Orderbook) TotalAmountAt(side types.Side, price float64) float64 {
	var amount float64
	o.sides[side].tree.Descend(func(item interfaces.Level) bool {
		amount += item.Amount()
		return item.Price() > price
	})

	return amount
}

func (o *Orderbook) Print() {
	print := ""
	var count int
	o.Iterate(types.Side__ASK, func(item interfaces.Level) bool {
		print = fmt.Sprintf("%f\t -- \t %f\n", item.Price(), item.Amount()) + print
		count++
		return count < 10
	})
	count = 0
	print += "---\n"
	o.Iterate(types.Side__BID, func(item interfaces.Level) bool {
		print += fmt.Sprintf("%f\t -- \t %f\n", item.Price(), item.Amount())
		count++
		return count < 10
	})

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Print(print)
}

type side struct {
	levels map[string]*level
	tree   *btree.BTreeG[interfaces.Level]
}

func newSide(s types.Side) *side {
	switch s {
	case types.Side__ASK:
		return &side{
			levels: make(map[string]*level),
			tree: btree.NewG(2, func(a, b interfaces.Level) bool {
				return a.Price() < b.Price()
			}),
		}
	case types.Side__BID:
		return &side{
			levels: make(map[string]*level),
			tree: btree.NewG(2, func(a, b interfaces.Level) bool {
				return a.Price() < b.Price()
			}),
		}
	default:
		return nil
	}
}
