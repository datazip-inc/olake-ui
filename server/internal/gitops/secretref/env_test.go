package secretref

import (
	"context"
	"errors"
	"testing"
)

func TestEnvProviderFetch(t *testing.T) {
	t.Setenv("PG_PASSWORD", "P@ss")

	testCases := []struct {
		name    string
		env     string
		want    string
		wantErr bool
	}{
		{
			name: "any variable",
			env:  "PG_PASSWORD",
			want: "P@ss",
		},
		{
			name:    "variable not set",
			env:     "PG_MISSING",
			wantErr: true,
		},
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
