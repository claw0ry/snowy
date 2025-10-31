// Copyright (c) 2025, Mads Moi-Aune <mads@moiaune.dev>
//
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type Client struct {
	InstanceURL string
	httpClient  *http.Client
}

func NewBasicAuthClient(c *BasicAuthCredential) *Client {
	return &Client{
		InstanceURL: c.InstanceURL,
		httpClient:  newBasicAuthClient(c.Username, c.Password),
	}
}

func NewOAuth2Client(ctx context.Context) (*Client, error) {
	profile, err := LoadOAuth2Profile()
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			return nil, fmt.Errorf("No profile detected. Please run '%s login'.", os.Args[0])
		}
		return nil, err
	}

	c, err := newOAuth2Client(ctx, profile)
	if err != nil {
		return nil, err
	}

	return &Client{
		InstanceURL: profile.InstanceURL,
		httpClient:  c,
	}, nil
}

func (c Client) Get(endpoint string, params url.Values) (*http.Response, error) {
	uri := fmt.Sprintf("%s/api/now/table/%s?%s", c.InstanceURL, endpoint, params.Encode())
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("[get] failed to create request: %w", err)
	}

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

	req.Header.Add("Content-Type", "application/json")

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

	req.Header.Add("Content-Type", "application/json")

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

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[delete] failed to get response: %w", err)
	}

	return res, nil
}
