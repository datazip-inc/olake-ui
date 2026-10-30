package reconciler

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/gitops/secretref"
)

type fakeProvider struct {
	secret string
	err    error
}

func (f *fakeProvider) Fetch(context.Context, string) (string, error) {
	return f.secret, f.err
}

func TestSecretRefsResolve(t *testing.T) {
	const withRef = `{"host":"pg","password":"${aws-sm:olake/pg}"}`

	testCases := []struct {
		name            string
		config          string
		provider        *fakeProvider
		want            string
		wantFound       bool
		wantErr         string
		wantNonRetrying bool
	}{
		{
			name:      "resolves a reference",
			config:    withRef,
			provider:  &fakeProvider{secret: "p"},
			want:      `{"host":"pg","password":"p"}`,
			wantFound: true,
		},
		{
			name:     "config without references",
			config:   `{"host":"pg"}`,
			provider: &fakeProvider{},
			want:     `{"host":"pg"}`,
		},
		{
			name:            "unresolvable reference is non-retryable",
			config:          withRef,
			provider:        &fakeProvider{err: errors.Join(secretref.ErrUnresolvable, errors.New("access denied"))},
			wantErr:         "access denied",
			wantNonRetrying: true,
		},
		{
			name:     "transient error is retried",
			config:   withRef,
			provider: &fakeProvider{err: errors.New("connection reset")},
			wantErr:  "connection reset",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			refs := &SecretRefs{
				resolver: secretref.NewWithProviders(map[string]secretref.Provider{"aws-sm": tc.provider}),
			}

			got, found, err := refs.resolve(context.Background(), tc.config)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want it to contain %q", err, tc.wantErr)
				}
				if errors.Is(err, ErrNonRetryable) != tc.wantNonRetrying {
					t.Fatalf("errors.Is(err, ErrNonRetryable) = %v, want %v", !tc.wantNonRetrying, tc.wantNonRetrying)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			if got != tc.want || found != tc.wantFound {
				t.Fatalf("got %s found=%v, want %s found=%v", got, found, tc.want, tc.wantFound)
			}
		})
	}
}

func TestSecretRefsMayRefresh(t *testing.T) {
	testCases := []struct {
		name    string
		refresh time.Duration
		config  string
		want    bool
	}{
		{name: "config with a reference", refresh: time.Hour, config: `{"p":"${aws-sm:x}"}`, want: true},
		{name: "config without a reference", refresh: time.Hour, config: `{"p":"plain"}`, want: false},
		{name: "refresh turned off", refresh: 0, config: `{"p":"${aws-sm:x}"}`, want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			refs := &SecretRefs{refresh: tc.refresh}
			if got := refs.mayRefresh(tc.config); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSecretRefsNext(t *testing.T) {
	testCases := []struct {
		name    string
		refresh time.Duration
		hasRefs bool
		wantMin time.Duration
		wantMax time.Duration
	}{
		{name: "resource with references is re-checked with jitter", refresh: time.Hour, hasRefs: true, wantMin: time.Hour, wantMax: time.Hour + 6*time.Minute},
		{name: "resource without references is not re-checked", refresh: time.Hour, hasRefs: false},
		{name: "refresh turned off", refresh: 0, hasRefs: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			refs := &SecretRefs{refresh: tc.refresh}
			got := refs.next(tc.hasRefs)
			if got < tc.wantMin || got > tc.wantMax {
				t.Fatalf("next = %s, want between %s and %s", got, tc.wantMin, tc.wantMax)
			}
		})
	}
}

func TestNewSecretRefs(t *testing.T) {
	testCases := []struct {
		name    string
		refresh time.Duration
		want    time.Duration
	}{
		{name: "interval is used as set", refresh: 15 * time.Minute, want: 15 * time.Minute},
		{name: "zero turns refresh off", refresh: 0, want: 0},
		{name: "value without a unit (60ns) is raised to the minimum", refresh: 60, want: minSecretRefresh},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			refs := NewSecretRefs(&appconfig.Config{GitOpsSecretsRefreshInterval: tc.refresh}, secretref.K8sConfig{})
			if refs.refresh != tc.want {
				t.Fatalf("refresh = %s, want %s", refs.refresh, tc.want)
			}
		})
	}
}
