package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/parnurzeal/gorequest"
)

type HttpClient struct {
	*gorequest.SuperAgent
}

func NewResut(method, targetUrl string, body interface{}) *HttpClient {
	client := &HttpClient{
		SuperAgent: gorequest.New(),
	}
	requestBody, _ := json.Marshal(body)
	logs.Debug("resquest body:", string(requestBody))
	client.SuperAgent.CustomMethod(method, targetUrl)
	if strings.ToUpper(method) == gorequest.POST && body != nil {
		client.SuperAgent.Send(body)
	}
	return client
}

func (c *HttpClient) Retry(retryerCount int, retryerTime time.Duration) *HttpClient {
	c.SuperAgent.Retry(retryerCount, retryerTime)
	return c
}

func (c *HttpClient) SetHeaders(headers map[string]string) *HttpClient {
	for k, v := range headers {
		c.Set(k, v)
	}
	return c
}
func (s *HttpClient) EndStruct(v interface{}, callback ...func(response gorequest.Response, v interface{}, body []byte, errs []error)) (gorequest.Response, []byte, []error) {
	response, body, errors := s.SuperAgent.EndStruct(v, callback...)
	logs.Debug(`response bode: `, response.Status)
	logs.Debug(`response body: `, string(body))
	return response, body, errors
}
