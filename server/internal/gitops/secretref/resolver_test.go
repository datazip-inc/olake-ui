package secretref

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeProvider struct {
	secrets map[string]string
	err     error
}

func (f *fakeProvider) Fetch(_ context.Context, path string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	secret, ok := f.secrets[path]
	if !ok {
		return "", unresolvable("secret %q not found", path)
	}
	return secret, nil
}

var testSecrets = map[string]string{
	"olake/pg":  `{"password":"P&ss<word>!123","port":5432,"ssl":true,"db_port":"5433"}`,
	"olake/raw": "Sup3rSecret",
}

func TestResolverResolve(t *testing.T) {
	testCases := []struct {
		name             string
		config           string
		providerErr      error
		want             string
		wantFound        bool
		wantErr          string
		wantUnresolvable bool
	}{
		{
			name:   "no references is returned unchanged",
			config: `{"host":"h", "port": 5432}`,
			want:   `{"host":"h", "port": 5432}`,
		},
		{
			name:      "key in JSON secret",
			config:    `{"host":"h","password":"${aws-sm:olake/pg#password}"}`,
			want:      `{"host":"h","password":"P\u0026ss\u003cword\u003e!123"}`,
			wantFound: true,
		},
		{
			name:      "whole secret in a nested object, numbers kept",
			config:    `{"ssh_config":{"private_key":"${aws-sm:olake/raw}"},"n":1.50}`,
			want:      `{"n":1.50,"ssh_config":{"private_key":"Sup3rSecret"}}`,
			wantFound: true,
		},
		{
			name:      "key value keeps its JSON type",
			config:    `{"port":"${aws-sm:olake/pg#port}","ssl":"${aws-sm:olake/pg#ssl}","db_port":"${aws-sm:olake/pg#db_port}"}`,
			want:      `{"db_port":"5433","port":5432,"ssl":true}`,
			wantFound: true,
		},
		{
			name:   "partial reference is a literal",
			config: `{"password":"x${aws-sm:olake/pg}","other":"${nocolon}"}`,
			want:   `{"other":"${nocolon}","password":"x${aws-sm:olake/pg}"}`,
		},
		{
			name:             "unsupported provider",
			config:           `{"p":"${vault:kv/pg}"}`,
			wantErr:          `unsupported secret provider "vault"`,
			wantUnresolvable: true,
		},
		{
			name:             "missing secret",
			config:           `{"p":"${aws-sm:olake/none}"}`,
			wantErr:          `secret "olake/none" not found`,
			wantUnresolvable: true,
		},
		{
			name:             "key on a secret that is not JSON",
			config:           `{"p":"${aws-sm:olake/raw#password}"}`,
			wantErr:          "not a JSON object",
			wantUnresolvable: true,
		},
		{
			name:             "missing key",
			config:           `{"p":"${aws-sm:olake/pg#user}"}`,
			wantErr:          `key "user" not found`,
			wantUnresolvable: true,
		},
		{
			name:        "transient provider error is passed on",
			config:      `{"p":"${aws-sm:olake/pg}"}`,
			providerErr: errors.New("connection reset"),
			wantErr:     "connection reset",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := &fakeProvider{secrets: testSecrets, err: tc.providerErr}
			r := NewWithProviders(map[string]Provider{"aws-sm": provider})

			got, found, err := r.Resolve(context.Background(), tc.config)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want it to contain %q", err, tc.wantErr)
				}
				if errors.Is(err, ErrUnresolvable) != tc.wantUnresolvable {
					t.Fatalf("errors.Is(err, ErrUnresolvable) = %v, want %v", !tc.wantUnresolvable, tc.wantUnresolvable)
				}
				if strings.Contains(err.Error(), "Sup3rSecret") {
					t.Fatalf("error leaks the secret value: %q", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if got != tc.want || found != tc.wantFound {
				t.Fatalf("got %s found=%v, want %s found=%v", got, found, tc.want, tc.wantFound)
			}
		})
	}
}
