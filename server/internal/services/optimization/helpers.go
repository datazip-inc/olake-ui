package optimization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

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

// (flexible parsing): performs an HTTP request and parses the response returning raw result as interface{}
func (c *Service) Do(ctx context.Context, method, path string, queryParams url.Values, body interface{}) (interface{}, error) {
	respBody, err := c.DoRequest(ctx, method, path, queryParams, body)
	if err != nil {
		return nil, err
	}

	resp, err := parseResponse(respBody)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %s", err)
	}

	return result, nil
}

// (type-safe parsing): performs an HTTP request and parses the response returning raw result into the provided result pointer
func (c *Service) DoInto(ctx context.Context, method, path string, queryParams url.Values, body, result interface{}) error {
	respBody, err := c.DoRequest(ctx, method, path, queryParams, body)
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
func (c *Service) DoExec(ctx context.Context, method, path string, queryParams url.Values, body interface{}) error {
	respBody, err := c.DoRequest(ctx, method, path, queryParams, body)
	if err != nil {
		return err
	}

	_, err = parseResponse(respBody)
	return err
}
