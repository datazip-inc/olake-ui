package gitops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitGitOps(t *testing.T) {
	ctx := context.Background()
	assert.NoError(t, InitGitOps(ctx, Config{Namespace: "olake"}, nil), "no mode means GitOps is off")
	// a typo in GITOPS_MODE must stop startup, not silently run without GitOps
	assert.ErrorContains(t, InitGitOps(ctx, Config{Mode: "flux", Namespace: "olake"}, nil), "unknown GITOPS_MODE flux")
}
