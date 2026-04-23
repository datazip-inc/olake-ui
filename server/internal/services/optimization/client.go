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
	"sync"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
)

type Service struct {
	baseURL  string
	username string
	password string
	client   *http.Client

	mu           sync.RWMutex // allow multiple read access, but only one write access
	apiKey       string
	apiSecret    string
	tokenVersion int64
}

// Token Expiration: There is "no" expiration logic for optimization
// https://github.com/datazip-inc/olake-fusion/blob/master/amoro-ams/src/main/java/org/apache/amoro/server/dashboard/controller/ApiTokenController.java
func NewClient() (*Service, error) {
	baseURL, username, password, err := getCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %s", err)
	}

	apiKey, apiSecret, err := generateToken(baseURL, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %s", err)
	}

	return &Service{
		baseURL:   baseURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		username:  username,
		password:  password,
		client: &http.Client{
			Timeout: constants.OptMaxTimeout,
		},
	}, nil
}

func (s *Service) credentials() (apiKey, apiSecret string, version int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.apiKey, s.apiSecret, s.tokenVersion
}

func (s *Service) tryRefreshToken(observedVersion int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tokenVersion != observedVersion {
		return nil // another thread refreshsed
	}
	apiKey, apiSecret, err := generateToken(s.baseURL, s.username, s.password)
	if err != nil {
		return fmt.Errorf("token refresh failed: %s", err)
	}
	s.apiKey = apiKey
	s.apiSecret = apiSecret
	s.tokenVersion++
	return nil
}

// for optimization authentication: calculating md5: apiKey + encryptString + secret
func (s *Service) calculateSignature(apiKey, apiSecret string, params url.Values) string {
	encryptString := s.generateEncryptString(params)
	plainText := fmt.Sprintf("%s%s%s", apiKey, encryptString, apiSecret)
	hash := md5.Sum([]byte(plainText))
	return hex.EncodeToString(hash[:])
}

func (s *Service) generateEncryptString(params url.Values) string {
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

func (s *Service) sendRequest(ctx context.Context, method, path string, queryParams url.Values, bodyBytes []byte) ([]byte, int, http.Header, error) {
	apiKey, apiSecret, _ := s.credentials()
	signature := s.calculateSignature(apiKey, apiSecret, queryParams)
	queryParams.Set("apiKey", apiKey)
	queryParams.Set("signature", signature)

	fullURL := s.baseURL + path
	if len(queryParams) > 0 {
		fullURL += "?" + queryParams.Encode()
	}

	var bodyReader io.Reader
	if bodyBytes != nil {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to create request to optimization: %s", err)
	}
	if bodyBytes != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to send request: %s", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, resp.Header, fmt.Errorf("failed to read response body: %s", err)
	}

	return respBody, resp.StatusCode, resp.Header, nil
}

func (s *Service) DoRequest(ctx context.Context, method, path string, queryParams url.Values, body interface{}) ([]byte, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %s", err)
		}
	}

	_, _, version := s.credentials()
	respBody, statusCode, _, err := s.sendRequest(ctx, method, path, queryParams, bodyBytes)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusUnauthorized {
		if err := s.tryRefreshToken(version); err != nil {
			return nil, err
		}
		respBody, statusCode, _, err = s.sendRequest(ctx, method, path, queryParams, bodyBytes)
		if err != nil {
			return nil, err
		}
	}

	if statusCode >= 400 {
		return nil, &HTTPError{StatusCode: statusCode, Body: respBody}
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

	tokenReq, err := http.NewRequest("POST", baseURL+"/api/ams/v1/api/token/create", http.NoBody)
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
	cfg := appconfig.Load()

	baseURL := cfg.OptimizationBaseURL
	if baseURL == "" {
		return "", "", "", fmt.Errorf("failed to get optimization base URL: OPTIMIZATION_BASE_URL environment variable not set")
	}

	username := cfg.OptimizationUsername
	if username == "" {
		return "", "", "", fmt.Errorf("failed to get optimization username: USERNAME environment variable not set")
	}

	password := cfg.OptimizationPassword
	if password == "" {
		return "", "", "", fmt.Errorf("failed to get optimization password: PASSWORD environment variable not set")
	}

	return baseURL, username, password, nil
}
