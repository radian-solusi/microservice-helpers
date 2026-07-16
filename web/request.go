package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	mu         sync.RWMutex
	lastStatus int
	lastHeader http.Header
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: http.DefaultClient,
	}
}

func (c *Client) SetBaseURL(url string) {
	c.mu.Lock()
	c.baseURL = strings.TrimRight(url, "/")
	c.mu.Unlock()
}
func (c *Client) BaseURL() string   { c.mu.RLock(); defer c.mu.RUnlock(); return c.baseURL }
func (c *Client) SetToken(t string) { c.mu.Lock(); c.token = t; c.mu.Unlock() }

func (c *Client) Do(ctx context.Context, method, path string, params any) ([]byte, error) {
	req, err := c.buildRequest(ctx, method, path, params)
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	c.mu.RUnlock()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	c.mu.Lock()
	c.lastStatus = resp.StatusCode
	c.lastHeader = resp.Header.Clone()
	c.mu.Unlock()

	return body, nil
}

func (c *Client) LastStatusCode() int     { c.mu.RLock(); defer c.mu.RUnlock(); return c.lastStatus }
func (c *Client) LastHeader() http.Header { c.mu.RLock(); defer c.mu.RUnlock(); return c.lastHeader }

func (c *Client) buildRequest(ctx context.Context, method, path string, params any) (*http.Request, error) {
	c.mu.RLock()
	base := c.baseURL
	c.mu.RUnlock()

	if params == nil {
		req, err := http.NewRequestWithContext(ctx, method, base+path, nil)
		if err != nil {
			return nil, fmt.Errorf("new request: %w", err)
		}
		return req, nil
	}

	switch method {
	case http.MethodGet, http.MethodDelete:
		m := flattenParams(params)
		if len(m) == 0 {
			req, err := http.NewRequestWithContext(ctx, method, base+path, nil)
			if err != nil {
				return nil, fmt.Errorf("new request: %w", err)
			}
			return req, nil
		}
		q := make(url.Values, len(m))
		for k, v := range m {
			q.Set(k, v)
		}
		u := base + path + "?" + q.Encode()
		req, err := http.NewRequestWithContext(ctx, method, u, nil)
		if err != nil {
			return nil, fmt.Errorf("new request: %w", err)
		}
		return req, nil
	default:
		body, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal params: %w", err)
		}
		req, err := http.NewRequestWithContext(ctx, method, base+path, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("new request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	}
}

// flattenParams converts params to a url.Values-like map, handling the
// subset of types the existing client sees: map[string]any, map[string]string,
// or a JSON-marshalled struct.
func flattenParams(params any) map[string]string {
	switch p := params.(type) {
	case map[string]any:
		m := make(map[string]string, len(p))
		for k, v := range p {
			switch vv := v.(type) {
			case string:
				if vv != "" {
					m[k] = vv
				}
			case int:
				m[k] = strconv.Itoa(vv)
			case []string:
				if len(vv) > 0 {
					m[k] = strings.Join(vv, ",")
				}
			}
		}
		return m
	case map[string]string:
		m := make(map[string]string, len(p))
		for k, v := range p {
			if v != "" {
				m[k] = v
			}
		}
		return m
	}
	return nil
}
