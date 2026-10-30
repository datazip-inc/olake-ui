package reconciler

import (
	"context"
	"strings"
	"unicode"

	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

const (
	indicatorSpawn  = "spawn"
	indicatorDelete = "delete"

	indicatorNameMax  = 63
	terminationLogMax = 4096
)

// StartIndicator asks Temporal to spawn or delete the failure indicator for r. It is best effort:
// the status already carries the error, so a failed start is only logged. A StatusSink's Indicator
// calls it.
func StartIndicator(ctx context.Context, t *temporal.Temporal, r *ResourceData, action, errMsg string) {
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

// indicatorName returns "<name>-olake-<kind>" truncated to 63 chars.
func indicatorName(name, kind string) string {
	suffix := "-olake-" + kind
	n := sanitizeDNS1123(name)
	if n == "" {
		n = "unnamed"
	}
	if len(n)+len(suffix) > indicatorNameMax {
		n = strings.TrimRight(n[:indicatorNameMax-len(suffix)], "-")
	}
	return n + suffix
}

func sanitizeDNS1123(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		ok := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-'
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
