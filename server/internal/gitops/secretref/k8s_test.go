package secretref

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewK8sProvider(t *testing.T) {
	testCases := []struct {
		name    string
		cfg     K8sConfig
		wantErr bool
	}{
		{name: "Kubernetes mode", cfg: K8sConfig{Client: fake.NewClientBuilder().Build(), Namespace: "olake"}},
		{name: "file mode has no client", cfg: K8sConfig{}, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newK8sProvider(context.Background(), Config{K8s: tc.cfg})
			if tc.wantErr != errors.Is(err, ErrUnresolvable) {
				t.Fatalf("err = %v, want unresolvable error: %v", err, tc.wantErr)
			}
		})
	}
}

func TestK8sProviderFetch(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "pg-creds", Namespace: "olake"},
		Data:       map[string][]byte{"username": []byte("sa"), "password": []byte("P@ss")},
	}
	provider := &k8sProvider{K8sConfig{
		Client:    fake.NewClientBuilder().WithObjects(secret).Build(),
		Namespace: "olake",
	}}

	testCases := []struct {
		name    string
		secret  string
		want    string
		wantErr bool
	}{
		{name: "Secret data as a JSON object", secret: "pg-creds", want: `{"password":"P@ss","username":"sa"}`},
		{name: "missing Secret", secret: "other", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := provider.Fetch(context.Background(), tc.secret)
			if tc.wantErr {
				if !errors.Is(err, ErrUnresolvable) {
					t.Fatalf("err = %v, want ErrUnresolvable", err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %s %v, want %s", got, err, tc.want)
			}
		})
	}
}
