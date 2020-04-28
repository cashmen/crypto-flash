package util

import "encoding/json"
import "net/http"
import "io"

type RestClient struct {
	tag string
	client *http.Client
}

func NewRestClient() *RestClient {
	return &RestClient{
		tag: "RestClient",
		client: &http.Client{},
	}
}
func (rc *RestClient) do(req *http.Request, v interface{}) {
	res, err := rc.client.Do(req)
	if err != nil {
		Error(rc.tag, err.Error())
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(v)
	if err != nil {
		Error(rc.tag, err.Error())
	}
}
func (rc *RestClient) Get(url string, header *http.Header, body io.Reader,
		v interface{}) {
	req, err := http.NewRequest("GET", url, body)
	if err != nil {
		Error(rc.tag, err.Error())
	}
	if header != nil {
		req.Header = *header
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	rc.do(req, v)
}
func (rc *RestClient) Post(url string, header *http.Header, body io.Reader, 
		v interface{}) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		Error(rc.tag, err.Error())
	}
	if header != nil {
		req.Header = *header
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	rc.do(req, v)
}
func (rc *RestClient) Delete(url string, header *http.Header, body io.Reader, 
		v interface{}) {
	req, err := http.NewRequest("DELETE", url, body)
	if err != nil {
		Error(rc.tag, err.Error())
	}
	if header != nil {
		req.Header = *header
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	rc.do(req, v)
}