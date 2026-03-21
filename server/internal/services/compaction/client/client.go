package client

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
)

type Compaction struct {
	baseURL   string
	apiKey    string
	apiSecret string
	client    *http.Client
}

// Token Expiration: There is "no" expiration logic for compaction
// https://github.com/datazip-inc/olake-fusion/blob/master/amoro-ams/src/main/java/org/apache/amoro/server/dashboard/controller/ApiTokenController.java
func NewClient() (*Compaction, error) {
	baseURL, username, password, err := getCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %s", err)
	}

	apiKey, apiSecret, err := generateToken(baseURL, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %s", err)
	}

	return &Compaction{
		baseURL:   baseURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// for compaction authentication: calculating md5: apiKey + encryptString + secret
func (c *Compaction) calculateSignature(params url.Values) string {
	encryptString := c.generateEncryptString(params)
	plainText := fmt.Sprintf("%s%s%s", c.apiKey, encryptString, c.apiSecret)
	hash := md5.Sum([]byte(plainText))
	signature := hex.EncodeToString(hash[:])

	return signature
}

func (c *Compaction) generateEncryptString(params url.Values) string {
	// Filter out apiKey and signature
	filtered := make(url.Values)
	for k, v := range params {
		if k != "apiKey" && k != "signature" {
			filtered[k] = v
		}
	}

	// Sort keys
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build param string
	var parts []string
	for _, k := range keys {
		values := filtered[k]
		if len(values) > 0 && values[0] != "" {
			decodedValue, err := url.QueryUnescape(values[0])
			if err != nil {
				decodedValue = values[0]
			}
			parts = append(parts, k+decodedValue)
		}
	}

	// If no valid parameters, return current date in yyyyMMdd format
	if len(parts) == 0 {
		return time.Now().UTC().Format("20060102")
	}

	return strings.Join(parts, "")
}

func (c *Compaction) DoRequest(ctx context.Context, method, path string, queryParams url.Values, body interface{}) ([]byte, error) {
	signature := c.calculateSignature(queryParams)
	queryParams.Set("apiKey", c.apiKey)
	queryParams.Set("signature", signature)

	fullURL := c.baseURL + path
	if len(queryParams) > 0 {
		fullURL += "?" + queryParams.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}

		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func generateToken(baseURL, username, password string) (string, string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	loginPayload := map[string]string{"user": username, "password": password}
	loginBody, err := json.Marshal(loginPayload)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal login payload: %w", err)
	}

	loginReq, err := http.NewRequest("POST", baseURL+"/api/ams/v1/login", bytes.NewReader(loginBody))
	if err != nil {
		return "", "", fmt.Errorf("failed to create login request: %w", err)
	}
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("X-Request-Source", "Web")

	loginResp, err := client.Do(loginReq)
	if err != nil {
		return "", "", fmt.Errorf("login failed: %w", err)
	}
	defer loginResp.Body.Close()

	cookies := loginResp.Cookies()

	tokenReq, err := http.NewRequest("POST", baseURL+"/api/ams/v1/api/token/create", nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create token request: %w", err)
	}
	tokenReq.Header.Set("Content-Type", "application/json")
	tokenReq.Header.Set("X-Request-Source", "Web")
	for _, cookie := range cookies {
		tokenReq.AddCookie(cookie)
	}

	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		return "", "", fmt.Errorf("token creation failed: %w", err)
	}
	defer tokenResp.Body.Close()

	var result struct {
		Result struct {
			APIKey string `json:"apikey"`
			Secret string `json:"secret"`
		} `json:"result"`
	}

	if err := json.NewDecoder(tokenResp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("failed to parse token response: %w", err)
	}

	return result.Result.APIKey, result.Result.Secret, nil
}

func getCredentials() (string, string, string, error) {
	baseURL, err := web.AppConfig.String(constants.ConfCompactionBaseURL)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get compaction base URL: %s", err)
	}

	username, err := web.AppConfig.String(constants.ConfCompactionUsername)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get compaction username creds: %s", err)
	}

	password, err := web.AppConfig.String(constants.ConfCompactionPassword)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get compaction password creds: %s", err)
	}

	return baseURL, username, password, nil
}
