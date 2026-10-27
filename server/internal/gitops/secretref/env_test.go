package secretref

import (
	"context"
	"errors"
	"testing"
)

func TestEnvProviderFetch(t *testing.T) {
	t.Setenv("GITOPS_SECRET_PG_PASSWORD", "P@ss")
	t.Setenv("OLAKE_SECRET_KEY", "server-key")

	testCases := []struct {
		name    string
		env     string
		want    string
		wantErr bool
	}{
		{name: "variable with the prefix", env: "GITOPS_SECRET_PG_PASSWORD", want: "P@ss"},
		{name: "variable without the prefix is refused", env: "OLAKE_SECRET_KEY", wantErr: true},
		{name: "variable not set", env: "GITOPS_SECRET_MISSING", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := envProvider{}.Fetch(context.Background(), tc.env)
			if tc.wantErr {
				if !errors.Is(err, ErrUnresolvable) {
					t.Fatalf("err = %v, want ErrUnresolvable", err)
				}
				if got != "" {
					t.Fatalf("got %q, want no value", got)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %q %v, want %q", got, err, tc.want)
			}
		})
	}
}
