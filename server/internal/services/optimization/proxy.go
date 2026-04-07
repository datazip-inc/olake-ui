package optimization

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

func (s *Service) PiggyBacking(ctx context.Context, method, path string, queryParams url.Values, body json.RawMessage) ([]byte, int, http.Header, error) {
	var bodyBytes []byte
	if len(body) > 0 {
		bodyBytes = []byte(body)
	}

	_, _, version := s.credentials()
	respBody, statusCode, headers, err := s.sendRequest(ctx, method, path, queryParams, bodyBytes)
	if err != nil {
		return nil, 0, nil, err
	}

	if statusCode == http.StatusUnauthorized {
		if err := s.tryRefreshToken(version); err != nil {
			return nil, 0, nil, err
		}
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
