package util

import "encoding/json"
import "bytes"

func getJSONBytes(s interface{}) []byte {
	b, err := json.Marshal(s)
	if err != nil {
		Error("Order", err.Error())
	}
	return b
}
func GetJSONString(s interface{}) string {
	return string(getJSONBytes(s))
}
func GetJSONBuffer(s interface{}) *bytes.Buffer {
	return bytes.NewBuffer(getJSONBytes(s))
}