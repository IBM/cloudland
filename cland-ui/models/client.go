package models

import (
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
)

type HttpClient struct {
	*gorequest.SuperAgent
}

func NewResut(method, targetUrl string, body interface{}) *HttpClient {
	client := &HttpClient{
		SuperAgent: gorequest.New(),
	}
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
