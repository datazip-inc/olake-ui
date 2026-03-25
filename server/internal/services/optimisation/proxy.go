package optimisation

import (
	"context"
	"encoding/json"
	"net/url"
)

// Proxy forwards a request to AMS and returns the upstream HTTP status and raw body.
func (s *Service) Proxy(ctx context.Context, method, path string, queryParams url.Values, body json.RawMessage) (httpStatus int, raw json.RawMessage, err error) {
	var bodyArg interface{}
	if len(body) > 0 {
		bodyArg = body
	}

	status, data, err := s.executeAMS(ctx, method, path, queryParams, bodyArg)
	if err != nil {
		return 0, nil, err
	}
	return status, json.RawMessage(data), nil
}
