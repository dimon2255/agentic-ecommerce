package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type QueryBuilder struct {
	client  *Client
	table   string
	params  url.Values
	method  string
	body    any
	headers map[string]string
	single  bool
}

func (c *Client) From(table string) *QueryBuilder {
	return &QueryBuilder{
		client:  c,
		table:   table,
		params:  url.Values{},
		method:  "GET",
		headers: map[string]string{},
	}
}

func (q *QueryBuilder) Select(columns string) *QueryBuilder {
	q.params.Set("select", columns)
	return q
}

func (q *QueryBuilder) Eq(column, value string) *QueryBuilder {
	q.params.Set(column, "eq."+value)
	return q
}

func (q *QueryBuilder) Is(column, value string) *QueryBuilder {
	q.params.Set(column, "is."+value)
	return q
}

func (q *QueryBuilder) In(column string, values []string) *QueryBuilder {
	joined := ""
	for i, v := range values {
		if i > 0 {
			joined += ","
		}
		joined += `"` + v + `"`
	}
	q.params.Set(column, "in.("+joined+")")
	return q
}

func (q *QueryBuilder) Order(column, direction string) *QueryBuilder {
	q.params.Set("order", column+"."+direction)
	return q
}

func (q *QueryBuilder) Limit(n int) *QueryBuilder {
	q.params.Set("limit", fmt.Sprintf("%d", n))
	return q
}

func (q *QueryBuilder) Offset(n int) *QueryBuilder {
	q.params.Set("offset", fmt.Sprintf("%d", n))
	return q
}

func (q *QueryBuilder) Single() *QueryBuilder {
	q.single = true
	q.headers["Accept"] = "application/vnd.pgrst.object+json"
	return q
}

func (q *QueryBuilder) Insert(data any) *QueryBuilder {
	q.method = "POST"
	q.body = data
	q.headers["Prefer"] = "return=representation"
	return q
}

func (q *QueryBuilder) Update(data any) *QueryBuilder {
	q.method = "PATCH"
	q.body = data
	q.headers["Prefer"] = "return=representation"
	return q
}

func (q *QueryBuilder) Delete() *QueryBuilder {
	q.method = "DELETE"
	q.headers["Prefer"] = "return=representation"
	return q
}

func (q *QueryBuilder) Execute(result any) error {
	reqURL := fmt.Sprintf("%s/rest/v1/%s", q.client.baseURL, q.table)
	if len(q.params) > 0 {
		reqURL += "?" + q.params.Encode()
	}

	var bodyReader io.Reader
	if q.body != nil {
		jsonBody, err := json.Marshal(q.body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(q.method, reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("apikey", q.client.apiKey)
	req.Header.Set("Authorization", "Bearer "+q.client.apiKey)
	req.Header.Set("Content-Type", "application/json")

	for k, v := range q.headers {
		req.Header.Set(k, v)
	}

	resp, err := q.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("supabase error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) RPC(functionName string, params any, result any) error {
	reqURL := fmt.Sprintf("%s/rest/v1/rpc/%s", c.baseURL, functionName)

	var bodyReader io.Reader
	if params != nil {
		jsonBody, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshal params: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest("POST", reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("supabase rpc error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}
