package util

type Order struct {
	Market string `json:"market"`
	Side string `json:"side"`
	Price float64 `json:"price"`
	Type string `json:"type"`
	Size float64 `json:"size"`
	ReduceOnly bool `json:"reduceOnly"`
	Ioc bool `json:"ioc"`
	PostOnly bool `json:"postOnly"`
	ClientId *string `json:"clientId"`
	// for conditional order
	// market: true or false
	// limit: false
	RetryUntilFilled bool `json:"retryUntilFilled"`

	// for stop and take profit
	TriggerPrice float64 `json:"triggerPrice"`
	// specified for limit, otherwise market
	OrderPrice float64 `json:"orderPrice"`

	// for trailing stop
	// negative for sell, positive for buy
	TrailValue float64 `json:"trailValue"` 
}
func (o *Order) CreateMap() map[string]interface{} {
	result := make(map[string]interface{})
	if o.Type == "limit" || o.Type == "market" {
		result["market"] = o.Market
		result["side"] = o.Side
		if o.Price > 0 {
			result["price"] = o.Price
		} else {
			result["price"] = nil
		}
		result["type"] = o.Type
		result["size"] = o.Size
		result["reduceOnly"] = o.ReduceOnly
		//result["ioc"] = o.Ioc
		//result["postOnly"] = o.PostOnly
		//result["clientID"] = o.ClientID
	} else if o.Type == "stop" || o.Type == "takeProfit" || 
			o.Type == "trailingStop" {
		result["market"] = o.Market
		result["side"] = o.Side
		result["size"] = o.Size
		result["type"] = o.Type
		result["reduceOnly"] = o.ReduceOnly
		result["retryUntilFilled"] = o.RetryUntilFilled
		if o.Type == "trailingStop" {
			result["trailValue"] = o.TrailValue
		} else {
			result["triggerPrice"] = o.TriggerPrice
			if o.OrderPrice > 0 {
				result["retryUntilFilled"] = false
				result["orderPrice"] = o.OrderPrice
			}
		}
	}
	return result
}