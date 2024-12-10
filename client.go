package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Credentials struct {
	InstanceURL string
	Username    string
	Password    string
}

type Client struct {
	*Credentials
	httpClient *http.Client
}

func NewClient(c *Credentials) Client {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	return Client{
		Credentials: c,
		httpClient:  client,
	}
}

func (c Client) Get(endpoint string, params url.Values) (*http.Response, error) {
	uri := fmt.Sprintf("%s/api/now/table/%s?%s", c.InstanceURL, endpoint, params.Encode())
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("[get] failed to create request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(c.Username, c.Password)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[get] failed to get response: %w", err)
	}

	return res, nil
}

func (c Client) Post(endpoint string, params url.Values, body []byte) (*http.Response, error) {
	uri := fmt.Sprintf("%s/api/now/table/%s?%s", c.InstanceURL, endpoint, params.Encode())
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("[post] failed to create request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(c.Username, c.Password)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[post] failed to get response: %w", err)
	}

	return res, nil
}

func (c Client) Patch(endpoint string, params url.Values, body []byte) (*http.Response, error) {
	uri := fmt.Sprintf("%s/api/now/table/%s?%s", c.InstanceURL, endpoint, params.Encode())
	req, err := http.NewRequest(http.MethodPatch, uri, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("[patch] failed to create request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(c.Username, c.Password)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[patch] failed to get response: %w", err)
	}

	return res, nil
}

func (c Client) Delete(endpoint string, params url.Values) (*http.Response, error) {
	uri := fmt.Sprintf("%s/api/now/table/%s?%s", c.InstanceURL, endpoint, params.Encode())
	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("[delete] failed to create request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(c.Username, c.Password)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[delete] failed to get response: %w", err)
	}

	return res, nil
}
