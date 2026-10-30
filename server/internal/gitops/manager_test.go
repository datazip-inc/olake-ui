package gitops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitGitOps(t *testing.T) {
	ctx := context.Background()
	assert.NoError(t, InitGitOps(ctx, "", testNamespace, nil), "no mode means GitOps is off")
	// a typo in GITOPS_MODE must stop startup, not silently run without GitOps
	assert.ErrorContains(t, InitGitOps(ctx, "flux", testNamespace, nil), "unknown GITOPS_MODE flux")
}
