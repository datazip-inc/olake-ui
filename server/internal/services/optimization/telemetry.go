package optimization

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
)

// SetInstallID hands Fusion the anonymous OLake install id. The UI owns the id and Fusion stores
// it, so both products report telemetry under one identity no matter how often either restarts.
func (s *Service) SetInstallID(ctx context.Context, installID string) error {
	body := map[string]string{"install_id": installID}
	if err := s.DoExec(ctx, http.MethodPut, constants.OptPathTelemetryInstallID, url.Values{}, body); err != nil {
		return fmt.Errorf("failed to set telemetry install id in optimization: %s", err)
	}
	return nil
}
