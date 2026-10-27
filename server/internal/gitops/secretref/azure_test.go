package secretref

import (
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

func TestAzureSecretName(t *testing.T) {
	testCases := []struct {
		path        string
		wantVault   string
		wantName    string
		wantVersion string
		wantErr     bool
	}{
		{path: "olake-kv/pg-password", wantVault: "olake-kv", wantName: "pg-password"},
		{path: "olake-kv/pg-password/0123abcd", wantVault: "olake-kv", wantName: "pg-password", wantVersion: "0123abcd"},
		{path: "pg-password", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			vault, name, version, err := azureSecretName(tc.path)
			if tc.wantErr {
				if !errors.Is(err, ErrUnresolvable) {
					t.Fatalf("err = %v, want ErrUnresolvable", err)
				}
				return
			}
			if err != nil || vault != tc.wantVault || name != tc.wantName || version != tc.wantVersion {
				t.Fatalf("got %q %q %q %v", vault, name, version, err)
			}
		})
	}
}

func TestClassifyAzureError(t *testing.T) {
	testCases := []struct {
		name             string
		err              error
		wantUnresolvable bool
	}{
		{name: "forbidden", err: &azcore.ResponseError{StatusCode: 403}, wantUnresolvable: true},
		{name: "secret not found", err: &azcore.ResponseError{StatusCode: 404}, wantUnresolvable: true},
		{name: "throttling is retried", err: &azcore.ResponseError{StatusCode: 429}},
		{name: "5xx is retried", err: &azcore.ResponseError{StatusCode: 503}},
		{name: "network or token failure is retried", err: errors.New("dial tcp: no such host")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyAzureError(tc.err)
			if errors.Is(got, ErrUnresolvable) != tc.wantUnresolvable {
				t.Fatalf("errors.Is(got, ErrUnresolvable) = %v, want %v", !tc.wantUnresolvable, tc.wantUnresolvable)
			}
		})
	}
}
