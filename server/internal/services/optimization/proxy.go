package optimization

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

func (s *Service) Proxy(ctx context.Context, method, path string, queryParams url.Values, body json.RawMessage) (json.RawMessage, error) {
	var bodyArg interface{}
	if len(body) > 0 {
		bodyArg = body
	}

	data, err := s.DoRequest(ctx, method, path, queryParams, bodyArg)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func (s *Service) ProxyWithHeaders(ctx context.Context, method, path string, queryParams url.Values, body json.RawMessage) ([]byte, http.Header, error) {
	var bodyBytes []byte
	if len(body) > 0 {
		bodyBytes = []byte(body)
	}

	respBody, statusCode, headers, err := s.sendRequest(ctx, method, path, queryParams, bodyBytes)
	if err != nil {
		return nil, nil, err
	}

	if statusCode == http.StatusUnauthorized {
		if err := s.refreshToken(); err != nil {
			return nil, nil, err
		}
		respBody, statusCode, headers, err = s.sendRequest(ctx, method, path, queryParams, bodyBytes)
		if err != nil {
			return nil, nil, err
		}
	}

	if statusCode >= 400 {
		return nil, headers, &httpError{StatusCode: statusCode, Body: respBody}
	}

	return respBody, headers, nil
}

type httpError struct {
	StatusCode int
	Body       []byte
}

func (e *httpError) Error() string {
	return string(e.Body)
}
