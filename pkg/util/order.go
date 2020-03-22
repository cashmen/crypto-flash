package util

import "encoding/json"
import "bytes"

type Order struct {
	Market string `json:"market"`
	Side string `json:"side"`
	Price *float64 `json:"price"`
	Type string `json:"type"`
	Size float64 `json:"size"`
	ReduceOnly bool `json:"reduceOnly"`
	Ioc bool `json:"ioc"`
	PostOnly bool `json:"postOnly"`
	ClientId *string `json:"clientId"`
}

func (o *Order) getJSONBytes() []byte {
	obj, err := json.Marshal(o)
	if err != nil {
		Error("Order", err.Error())
	}
	return obj
}
func (o *Order) GetJSONString() string {
	return string(o.getJSONBytes())
}
func (o *Order) GetBuffer() *bytes.Buffer {
	return bytes.NewBuffer(o.getJSONBytes())
}