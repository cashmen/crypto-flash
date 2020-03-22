package util

type Row struct {
	Price float64
	Size float64
}
type Orderbook struct {
	Bid []Row
	Ask []Row
}
func (ob *Orderbook) Add(side string, price, size float64) {
	row := Row{ Price: price, Size: size }
	if side == "bid" {
		ob.Bid = append(ob.Bid, row)
	} else if side == "ask" {
		ob.Ask = append(ob.Ask, row)
	}
}