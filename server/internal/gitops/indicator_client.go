package gitops

import (
	"context"

	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

func spawnIndicatorViaTemporal(ctx context.Context, t *temporal.Temporal, r ResourceData, errMsg string) error {
	if t == nil {
		return nil
	}
	req := temporal.IndicatorRequest{
		Action:    "spawn",
		Name:      indicatorName(r.Name, r.Kind),
		Namespace: r.Namespace,
		Kind:      r.Kind,
		CRName:    r.Name,
		Message:   truncate(errMsg, terminationLogMax),
	}
	if err := t.StartIndicator(ctx, req); err != nil {
		logger.Errorf("start indicator workflow spawn %s/%s: %s", r.Namespace, r.Name, err)
	}
	return nil
}

func deleteIndicatorViaTemporal(ctx context.Context, t *temporal.Temporal, r ResourceData) error {
	if t == nil {
		return nil
	}
	req := temporal.IndicatorRequest{
		Action:    "delete",
		Name:      indicatorName(r.Name, r.Kind),
		Namespace: r.Namespace,
		Kind:      r.Kind,
		CRName:    r.Name,
	}
	if err := t.StartIndicator(ctx, req); err != nil {
		logger.Errorf("start indicator workflow delete %s/%s: %s", r.Namespace, r.Name, err)
	}
	return nil
}
