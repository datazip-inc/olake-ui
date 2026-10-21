package gitops

import (
	"context"

	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

const (
	indicatorSpawn  = "spawn"
	indicatorDelete = "delete"
)

// Indicator asks Temporal to spawn or delete the failure indicator for r.
func (s *K8sSink) Indicator(ctx context.Context, r *ResourceData, action, errMsg string) {
	startIndicator(ctx, s.temporal, r, action, errMsg)
}

func startIndicator(ctx context.Context, t *temporal.Temporal, r *ResourceData, action, errMsg string) {
	if t == nil {
		return
	}
	req := temporal.IndicatorRequest{
		Action:       action,
		Name:         indicatorName(r.Name, r.Kind),
		Namespace:    r.Namespace,
		Kind:         r.Kind,
		ResourceName: r.Name,
		Message:      truncate(errMsg, terminationLogMax),
	}
	if err := t.StartIndicator(ctx, &req); err != nil {
		logger.Errorf("start indicator workflow %s %s/%s: %s", action, r.Namespace, r.Name, err)
	}
}
