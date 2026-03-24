package optimisation

import (
	"context"
	"encoding/json"
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
