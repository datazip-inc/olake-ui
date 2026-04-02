package optimization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func (s *Service) ProxyWithHeaders(ctx context.Context, method, path string, queryParams url.Values, body json.RawMessage) ([]byte, int, http.Header, error) {
	var bodyBytes []byte
	if len(body) > 0 {
		bodyBytes = []byte(body)
	}

	respBody, statusCode, headers, err := s.sendRequest(ctx, method, path, queryParams, bodyBytes)
	if err != nil {
		return nil, 0, nil, err
	}

	if statusCode == http.StatusUnauthorized {
		var apiKey string
		var apiSecret string
		var err error
		apiKey, apiSecret, err = refreshToken(s.baseURL, s.username, s.password)
		if err != nil {
			return nil, 0, nil, fmt.Errorf("token refresh failed: %s", err)
		}
		s.apiKey = apiKey
		s.apiSecret = apiSecret

		respBody, statusCode, headers, err = s.sendRequest(ctx, method, path, queryParams, bodyBytes)
		if err != nil {
			return nil, 0, nil, err
		}
	}

	if statusCode >= 400 {
		return nil, statusCode, headers, &HTTPError{StatusCode: statusCode, Body: respBody}
	}

	return respBody, statusCode, headers, nil
}

type HTTPError struct {
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	return string(e.Body)
}
