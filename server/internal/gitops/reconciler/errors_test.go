package reconciler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNonRetryableError(t *testing.T) {
	assert.NoError(t, NonRetryableError(nil))

	cause := errors.New("bad spec")
	err := NonRetryableError(cause)
	assert.ErrorIs(t, err, ErrNonRetryable)
	assert.ErrorIs(t, err, cause)

	// wrapping twice must not repeat the marker in the message users see in olake.io/message
	assert.Equal(t, err, NonRetryableError(err))
}

func TestRequireSpec(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		userID    string
		wantID    int
		wantErr   string
	}{
		{name: "valid", projectID: "123", userID: "7", wantID: 7},
		{name: "missing project", projectID: "", userID: "7", wantErr: "data.projectId is required"},
		{name: "missing user", projectID: "123", userID: "", wantErr: "data.userId must be a positive integer"},
		{name: "non-numeric user", projectID: "123", userID: "admin", wantErr: "data.userId must be a positive integer"},
		{name: "zero user", projectID: "123", userID: "0", wantErr: "data.userId must be a positive integer"},
		{name: "negative user", projectID: "123", userID: "-1", wantErr: "data.userId must be a positive integer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := requireSpec(tt.projectID, tt.userID)
			if tt.wantErr != "" {
				assert.ErrorIs(t, err, ErrNonRetryable)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
		})
	}
}
