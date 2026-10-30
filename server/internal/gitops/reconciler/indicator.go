package reconciler

import (
	"context"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

const (
	indicatorSpawn  = "spawn"
	indicatorDelete = "delete"

	indicatorNameMax = 63
)

// StartIndicator asks Temporal to spawn or delete the failure indicator for r. It is best effort:
// the status already carries the error, so a failed start is only logged.
func StartIndicator(ctx context.Context, t *temporal.Temporal, r *ResourceData, action, errMsg string) {
	req := temporal.IndicatorRequest{
		Action:       action,
		Name:         indicatorName(r.Name, r.Kind),
		Namespace:    r.Namespace,
		Kind:         r.Kind,
		ResourceName: r.Name,
		Message:      errMsg,
	}
	if err := t.StartIndicator(ctx, &req); err != nil {
		logger.Errorf("start indicator workflow %s %s/%s: %s", action, r.Namespace, r.Name, err)
	}
}

// indicatorName returns "olake-failed-<kind>-<name>", the indicator pod or container name: it reads
// as a failure next to the other OLake pods, and the name is cut so the whole fits in 63 chars.
func indicatorName(name, kind string) string {
	prefix := "olake-failed-" + kind + "-"
	n := sanitizeDNS1123(name)
	if n == "" {
		n = "unnamed"
	}
	if len(prefix)+len(n) > indicatorNameMax {
		n = strings.TrimRight(n[:indicatorNameMax-len(prefix)], "-")
	}
	return prefix + n
}

func sanitizeDNS1123(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		// DNS-1123 allows ASCII only: unicode.IsLetter would let 'é' through and the pod would be rejected
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
		if !ok {
			r = '-'
		}
		if r == '-' {
			if prevDash {
				continue
			}
			prevDash = true
		} else {
			prevDash = false
		}
		b.WriteRune(r)
	}
	return strings.Trim(b.String(), "-")
}
