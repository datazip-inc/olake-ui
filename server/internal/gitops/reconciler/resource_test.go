package reconciler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResourceID(t *testing.T) {
	for ref, want := range map[string]int{"12": 12, "1": 1} {
		id, ok := parseResourceID(ref)
		assert.True(t, ok, ref)
		assert.Equal(t, want, id)
	}
	// anything else is a name, including names that start with digits
	for _, ref := range []string{"0", "-3", "12a", "pg", "", " 12"} {
		_, ok := parseResourceID(ref)
		assert.False(t, ok, ref)
	}
}
