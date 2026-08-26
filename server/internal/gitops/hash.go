package gitops

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

// ContentHash returns a stable SHA256 hex of the string map (keys sorted).
// Used for all four ConfigMap kinds (Source, Destination, Job, Streams).
func ContentHash(data map[string]string) string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, k := range keys {
		fmt.Fprintf(h, "%s=%s\n", k, data[k])
	}
	return hex.EncodeToString(h.Sum(nil))
}

// TODO: add support for Secrets — add bytesData(in map[string][]byte) map[string]string
//       to convert Secret.Data before passing to ContentHash.
