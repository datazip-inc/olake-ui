package optimization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// extracts the body-level status code.
func parseResponse(respBody []byte) (*dto.OptimizationResponse, error) {
	var resp dto.OptimizationResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %s", err)
	}
	if resp.Code != 200 {
		return nil, &HTTPError{StatusCode: resp.Code, Body: []byte(resp.Message)}
	}
	return &resp, nil
}

// (type-safe parsing): performs an HTTP request and parses the response returning raw result into the provided result pointer
func (s *Service) DoInto(ctx context.Context, method, path string, queryParams url.Values, body, result interface{}) error {
	respBody, err := s.DoRequest(ctx, method, path, queryParams, body)
	if err != nil {
		return err
	}

	resp, err := parseResponse(respBody)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp.Result, result); err != nil {
		return fmt.Errorf("failed to parse result: %s", err)
	}

	return nil
}

// (no-result): performs an HTTP request expecting no result payload, only checks for success
func (s *Service) DoExec(ctx context.Context, method, path string, queryParams url.Values, body interface{}) error {
	respBody, err := s.DoRequest(ctx, method, path, queryParams, body)
	if err != nil {
		return err
	}

	_, err = parseResponse(respBody)
	return err
}
