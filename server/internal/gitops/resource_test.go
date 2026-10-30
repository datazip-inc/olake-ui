package gitops

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndicatorName(t *testing.T) {
	tests := []struct {
		name string
		kind string
		want string
	}{
		{name: "orders", kind: KindJob, want: "orders-olake-job"},
		{name: "Prod_Orders.v2", kind: KindJob, want: "prod-orders-v2-olake-job"},
		{name: "a--b__c", kind: KindSource, want: "a-b-c-olake-source"},
		{name: "-edge-", kind: KindSource, want: "edge-olake-source"},
		{name: "___", kind: KindDestination, want: "unnamed-olake-destination"},
		{name: "", kind: KindStreams, want: "unnamed-olake-streams"},
		// cut to 63 chars without leaving a dash before the suffix
		{name: strings.Repeat("a", 52) + "-bbbbbbbbb", kind: KindJob, want: strings.Repeat("a", 52) + "-olake-job"},
		{name: strings.Repeat("x", 80), kind: KindJob, want: strings.Repeat("x", 53) + "-olake-job"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, indicatorName(tt.name, tt.kind))
		})
	}
}

func TestParseResourceID(t *testing.T) {
	for ref, want := range map[string]int{"12": 12, "1": 1} {
		id, ok := parseResourceID(ref)
		assert.True(t, ok, ref)
		assert.Equal(t, want, id)
	}
	// anything else is a name, including names that start with digits
	for _, ref := range []string{"0", "-3", "12a", "pg", "", " 12"} {
		_, ok := parseResourceID(ref)
		assert.False(t, ok, ref)
	}
}
