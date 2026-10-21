package gitops

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// ContentHash returns a stable SHA256 hex of the string map.
// encoding/json sorts map keys and escapes values, so distinct maps never
// serialise to the same bytes.
func ContentHash(data map[string]string) string {
	b, _ := json.Marshal(data)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
