package pan123

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	APIBaseURL   = "https://open-api.123pan.com"
	PlatformName = "open_platform"
)

type Client struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string

	mu        sync.RWMutex
	token     string
	expiredAt time.Time
}

func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (c *Client) getToken() (string, error) {
	c.mu.RLock()
	// refresh 5 minutes before expiry to avoid mid-request expiration
	if c.token != "" && time.Now().Before(c.expiredAt.Add(-5*time.Minute)) {
		t := c.token
		c.mu.RUnlock()
		return t, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Now().Before(c.expiredAt.Add(-5*time.Minute)) {
		return c.token, nil
	}

	reqBody := AccessTokenReq{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
	}
	b, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", APIBaseURL+"/api/v1/access_token", bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("create token request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Platform", PlatformName)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("token request returned http status %d", resp.StatusCode)
	}
	var apiResp BaseResp
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("decode token response error: %w", err)
	}

	if apiResp.Code != 0 {
		return "", fmt.Errorf("123pan API error (get token), code: %d, msg: %s", apiResp.Code, apiResp.Message)
	}

	var data AccessTokenRespData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return "", fmt.Errorf("decode AccessTokenRespData error: %w", err)
	}
	if data.AccessToken == "" {
		return "", fmt.Errorf("123pan returned an empty access token")
	}

	exp, err := time.Parse(time.RFC3339, data.ExpiredAt)
	if err != nil {
		exp = time.Now().Add(2 * time.Hour)
	}

	c.token = data.AccessToken
	c.expiredAt = exp
	return c.token, nil
}

func (c *Client) DoJSONRequest(method, url string, body interface{}, target interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body error: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	token, err := c.getToken()
	if err != nil {
		return fmt.Errorf("获取 123pan AccessToken 失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Platform", PlatformName)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		snippet := string(bytes.TrimSpace(respBytes))
		if snippet != "" {
			return fmt.Errorf("http status %d: %s", resp.StatusCode, snippet)
		}
		return fmt.Errorf("http status %d", resp.StatusCode)
	}

	if target != nil {
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response body error: %w", err)
		}
		if err := json.Unmarshal(respBytes, target); err != nil {
			return fmt.Errorf("unmarshal response error: %w", err)
		}
	}

	return nil
}

// DoRawPUT uploads binary data to a presigned URL without auth headers (required by 123pan).
func (c *Client) DoRawPUT(presignedURL string, data io.Reader, size int64) error {
	req, err := http.NewRequest("PUT", presignedURL, data)
	if err != nil {
		return fmt.Errorf("invalid presigned upload URL")
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = size

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Transport and redirect errors can include the signed URL.
		return fmt.Errorf("PUT request to %s failed", req.URL.Hostname())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("PUT http status %d", resp.StatusCode)
	}

	return nil
}
