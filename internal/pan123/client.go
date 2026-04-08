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
	APIBaseURL    = "https://open-api.123pan.com"
	UploadBaseURL = "https://open-api.123pan.com"
	PlatformName  = "open_platform"
)

type Client struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	
	mu        sync.RWMutex
	token     string
	expiredAt time.Time
}

// NewClient 初始化 123pan API 的定制化客户端
func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // 设置超时防阻塞
		},
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// getToken 安全获取当前有效的 token，必要时自动刷新并缓存
func (c *Client) getToken() (string, error) {
	c.mu.RLock()
	// 提前 5 分钟刷新以防恰好在请求过程中过期
	if c.token != "" && time.Now().Before(c.expiredAt.Add(-5*time.Minute)) {
		t := c.token
		c.mu.RUnlock()
		return t, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	// 双重检查锁定
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
		return "", fmt.Errorf("http execute token request failed: %w", err)
	}
	defer resp.Body.Close()

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

	exp, err := time.Parse(time.RFC3339, data.ExpiredAt)
	if err != nil {
		// 解析失败时给个保守过期时间
		exp = time.Now().Add(2 * time.Hour)
	}

	c.token = data.AccessToken
	c.expiredAt = exp
	return c.token, nil
}

// DoJSONRequest 发起基于 JSON 的标准请求，并自动装配所需的三大件 Header
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

	// 123pan 开放平台标准 Header 规范
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Platform", PlatformName)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body error: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http status code error %d, body: %s", resp.StatusCode, string(respBytes))
	}

	// 反序列化 123pan 返回的整体数据 (BaseResp 格式需在调用该处的人自己指定或由外包装封入)
	if target != nil {
		if err := json.Unmarshal(respBytes, target); err != nil {
			return fmt.Errorf("unmarshal response error: %w, payload: %s", err, string(respBytes))
		}
	}

	return nil
}

// DoRawPUT 对于实际的分片数据上传，需要剥离 Auth 头部并发起干净的 PUT 请求
func (c *Client) DoRawPUT(url string, data io.Reader, size int64) error {
	req, err := http.NewRequest("PUT", url, data)
	if err != nil {
		return fmt.Errorf("create put request error: %w", err)
	}

	// 对于纯二进制上传必须准确标注 Content-Type 和 Length
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = size 

	// 文档强调: PUT请求的header中请不要携带Authorization、Platform参数
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http put execute error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http error code %d during PUT upload, body: %s", resp.StatusCode, string(respBytes))
	}

	return nil
}
