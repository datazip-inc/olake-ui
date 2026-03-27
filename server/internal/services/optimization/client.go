package optimization

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

type Service struct {
	baseURL   string
	apiKey    string
	apiSecret string
	username  string
	password  string
	client    *http.Client
}

// Token Expiration: There is "no" expiration logic for optimization
// https://github.com/datazip-inc/olake-fusion/blob/master/amoro-ams/src/main/java/org/apache/amoro/server/dashboard/controller/ApiTokenController.java
func NewClient() (*Service, error) {
	baseURL, username, password, err := getCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %s", err)
	}

	apiKey, apiSecret, err := generateToken(baseURL, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %s", err)
	}

	return &Service{
		baseURL:   baseURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		username:  username,
		password:  password,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *Service) refreshToken() error {
	apiKey, apiSecret, err := generateToken(c.baseURL, c.username, c.password)
	if err != nil {
		return err
	}
	c.apiKey = apiKey
	c.apiSecret = apiSecret
	return nil
}

// for optimization authentication: calculating md5: apiKey + encryptString + secret
func (c *Service) calculateSignature(params url.Values) string {
	encryptString := c.generateEncryptString(params)
	plainText := fmt.Sprintf("%s%s%s", c.apiKey, encryptString, c.apiSecret)
	hash := md5.Sum([]byte(plainText))
	signature := hex.EncodeToString(hash[:])

	return signature
}

func (c *Service) generateEncryptString(params url.Values) string {
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

func (c *Service) sendRequest(ctx context.Context, method, path string, queryParams url.Values, bodyBytes []byte) ([]byte, int, error) {
	signature := c.calculateSignature(queryParams)
	queryParams.Set("apiKey", c.apiKey)
	queryParams.Set("signature", signature)

	fullURL := c.baseURL + path
	if len(queryParams) > 0 {
		fullURL += "?" + queryParams.Encode()
	}

	var bodyReader io.Reader
	if bodyBytes != nil {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %s", err)
	}
	if bodyBytes != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request: %s", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %s", err)
	}

	return respBody, resp.StatusCode, nil
}

func (c *Service) DoRequest(ctx context.Context, method, path string, queryParams url.Values, body interface{}) ([]byte, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %s", err)
		}
	}

	respBody, statusCode, err := c.sendRequest(ctx, method, path, queryParams, bodyBytes)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusUnauthorized {
		if err := c.refreshToken(); err != nil {
			return nil, fmt.Errorf("token refresh failed: %s", err)
		}
		respBody, statusCode, err = c.sendRequest(ctx, method, path, queryParams, bodyBytes)
		if err != nil {
			return nil, err
		}
	}

	if statusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", statusCode, string(respBody))
	}

	return respBody, nil
}

func generateToken(baseURL, username, password string) (string, string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	loginPayload := map[string]string{"user": username, "password": password}
	loginBody, err := json.Marshal(loginPayload)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal login payload: %s", err)
	}

	loginReq, err := http.NewRequest("POST", baseURL+"/api/ams/v1/login", bytes.NewReader(loginBody))
	if err != nil {
		return "", "", fmt.Errorf("failed to create login request: %s", err)
	}
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("X-Request-Source", "Web")

	loginResp, err := client.Do(loginReq)
	if err != nil {
		return "", "", fmt.Errorf("login failed: %s", err)
	}
	defer loginResp.Body.Close()

	cookies := loginResp.Cookies()

	tokenReq, err := http.NewRequest("POST", baseURL+"/api/ams/v1/api/token/create", nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create token request: %s", err)
	}
	tokenReq.Header.Set("Content-Type", "application/json")
	tokenReq.Header.Set("X-Request-Source", "Web")
	for _, cookie := range cookies {
		tokenReq.AddCookie(cookie)
	}

	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		return "", "", fmt.Errorf("token creation failed: %s", err)
	}
	defer tokenResp.Body.Close()

	var result struct {
		Result struct {
			APIKey string `json:"apikey"`
			Secret string `json:"secret"`
		} `json:"result"`
	}

	if err := json.NewDecoder(tokenResp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("failed to parse token response: %s", err)
	}

	return result.Result.APIKey, result.Result.Secret, nil
}

func getCredentials() (string, string, string, error) {
	baseURL, err := web.AppConfig.String(constants.ConfOptimizationBaseURL)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get optimization base URL: %s", err)
	}

	username, err := web.AppConfig.String(constants.ConfOptimizationUsername)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get optimization username creds: %s", err)
	}

	password, err := web.AppConfig.String(constants.ConfOptimizationPassword)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get optimization password creds: %s", err)
	}

	return baseURL, username, password, nil
}
