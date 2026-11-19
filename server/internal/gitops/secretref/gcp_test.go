package secretref

import (
	"errors"
	"testing"
)

func TestGCPVersionName(t *testing.T) {
	testCases := []struct {
		path    string
		project string
		want    string
		wantErr bool
	}{
		{
			path:    "pg",
			project: "p1",
			want:    "projects/p1/secrets/pg/versions/latest",
		},
		{
			path: "projects/p2/secrets/pg",
			want: "projects/p2/secrets/pg/versions/latest",
		},
		{
			path: "projects/p2/secrets/pg/versions/3",
			want: "projects/p2/secrets/pg/versions/3",
		},
		{
			path:    "pg",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.path+"@"+tc.project, func(t *testing.T) {
			got, err := gcpVersionName(tc.path, tc.project)
			if tc.wantErr {
				if !errors.Is(err, ErrUnresolvable) {
					t.Fatalf("err = %v, want ErrUnresolvable", err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %q %v, want %q", got, err, tc.want)
			}
		})
	}
}
