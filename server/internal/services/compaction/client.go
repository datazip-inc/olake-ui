package compaction

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Compaction struct {
	baseURL   string
	apiKey    string
	apiSecret string
	client    *http.Client
}

func NewClient(baseURL, apiKey, apiSecret string) *Compaction {
	// Use provided values, fallback to environment variables, then to defaults
	return &Compaction{
		baseURL:   baseURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Compaction) calculateSignature(params url.Values) string {
	// Generate encrypt string from params (matching Java implementation)
	encryptString := c.generateEncryptString(params)

	// calculating md5: apiKey + encryptString + secret
	plainText := fmt.Sprintf("%s%s%s", c.apiKey, encryptString, c.apiSecret)
	hash := md5.Sum([]byte(plainText))
	signature := hex.EncodeToString(hash[:])

	log.Printf("[DEBUG] Signature calculation: encryptString=%s, plainText=%s, signature=%s", encryptString, plainText, signature)

	return signature
}

// generateEncryptString matches Java's ParamSignatureCalculator logic
func (c *Compaction) generateEncryptString(params url.Values) string {
	// Remove apiKey and signature from params
	filtered := make(url.Values)
	for k, v := range params {
		if k != "apiKey" && k != "signature" {
			filtered[k] = v
		}
	}

	// If no params left, return current date in yyyyMMdd format
	if len(filtered) == 0 {
		return time.Now().UTC().Format("20060102")
	}

	// Sort keys and concatenate as key+value
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		values := filtered[k]
		if len(values) > 0 && values[0] != "" {
			// URL decode the value (matching Java's URLDecoder.decode)
			decodedValue, err := url.QueryUnescape(values[0])
			if err != nil {
				decodedValue = values[0]
			}
			parts = append(parts, k+decodedValue)
		}
	}

	// If all values were empty, return date
	if len(parts) == 0 {
		return time.Now().UTC().Format("20060102")
	}

	return strings.Join(parts, "")
}

func (c *Compaction) doRequest(ctx context.Context, method, path string, queryParams url.Values, body interface{}) ([]byte, error) {
	// Calculate signature BEFORE adding apiKey to params
	signature := c.calculateSignature(queryParams)
	queryParams.Set("apiKey", c.apiKey)
	queryParams.Set("signature", signature)

	fullURL := c.baseURL + path
	if len(queryParams) > 0 {
		fullURL += "?" + queryParams.Encode()
	}

	log.Printf("[DEBUG] Full request URL: %s", fullURL)

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
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
