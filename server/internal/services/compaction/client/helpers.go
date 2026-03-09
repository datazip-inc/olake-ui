package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// (flexible parsing): performs an HTTP request and parses the response returning raw result as interface{}
func (c *Compaction) Do(ctx context.Context, method, path string, queryParams url.Values, body interface{}) (interface{}, error) {
	respBody, err := c.DoRequest(ctx, method, path, queryParams, body)
	if err != nil {
		return nil, err
	}

	var resp dto.CompactionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %s", err)
	}

	if resp.Code != 200 {
		return nil, fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	var result interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %s", err)
	}

	return result, nil
}

// (type-safe parsing): performs an HTTP request and parses the response returning raw result into the provided result pointer
func (c *Compaction) DoInto(ctx context.Context, method, path string, queryParams url.Values, body interface{}, result interface{}) error {
	respBody, err := c.DoRequest(ctx, method, path, queryParams, body)
	if err != nil {
		return err
	}

	var resp dto.CompactionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %s", err)
	}

	if resp.Code != 200 {
		return fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	if err := json.Unmarshal(resp.Result, result); err != nil {
		return fmt.Errorf("failed to parse result: %w", err)
	}

	return nil
}

// (no-result data): performs an HTTP request and validates the response
func (c *Compaction) DoAndValidate(ctx context.Context, method, path string, queryParams url.Values, body interface{}) error {
	respBody, err := c.DoRequest(ctx, method, path, queryParams, body)
	if err != nil {
		return err
	}

	var resp dto.CompactionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Code != 200 {
		return fmt.Errorf("fusion error (code %d): %s", resp.Code, resp.Message)
	}

	return nil
}
