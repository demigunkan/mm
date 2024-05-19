package interfaces

type Level interface {
	Amount() float64
	Price() float64
	NetPrice() float64
}
