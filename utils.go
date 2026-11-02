package tls_client

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"strings"
)

func Int64ToInt(x int64) (int, error) {
	if x < math.MinInt || x > math.MaxInt {
		return 0, fmt.Errorf("int64 value %d out of int range [%d, %d]", x, math.MinInt, math.MaxInt)
	}
	return int(x), nil
}

// generateGREASESettingID generates a valid GREASE setting ID
// GREASE IDs are of the form 0x1f * N + 0x21 where N is random
// Chrome uses very large N values, producing setting IDs like 57836956465
func generateGREASESettingID() uint64 {
	// Generate large N values similar to Chrome (produces 10-11 digit IDs)
	// N between 1,000,000,000 and 10,000,000,000
	nBig, _ := rand.Int(rand.Reader, big.NewInt(9000000000))
	n := uint64(1000000000) + nBig.Uint64()
	return 0x1f*n + 0x21
}

// generateGREASESettingValue generates a random non-zero 32-bit value for GREASE
func generateGREASESettingValue() uint64 {
	var buf [4]byte
	rand.Read(buf[:])
	val := binary.BigEndian.Uint32(buf[:])
	// Chrome never sends 0
	if val == 0 {
		val = 1
	}
	return uint64(val)
}

// sanitizeRawQuery percent-encodes every byte that RFC 3986 does not allow
// raw in a URI query (query = pchar / "/" / "?", where pchar covers
// pct-encoded triplets). url.Parse keeps such bytes verbatim in RawQuery and
// both the HTTP/1 request line and the HTTP/2 :path carry RawQuery verbatim,
// so a literal space for example would break the request line. Browsers never
// send these bytes unencoded.
func sanitizeRawQuery(query string) string {
	for i := 0; i < len(query); i++ {
		if !isValidRawQueryByte(query[i]) {
			var b strings.Builder
			b.Grow(len(query) + 8)
			b.WriteString(query[:i])

			for j := i; j < len(query); j++ {
				if isValidRawQueryByte(query[j]) {
					b.WriteByte(query[j])
				} else {
					fmt.Fprintf(&b, "%%%02X", query[j])
				}
			}

			return b.String()
		}
	}

	return query
}

// isValidRawQueryByte reports whether b may appear unencoded in a URI query.
// '%' counts as valid: it introduces a pct-encoded triplet, and encoding it
// again would double-encode already encoded input.
func isValidRawQueryByte(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9':
		return true
	}

	switch b {
	case '-', '.', '_', '~', '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', ':', '@', '/', '?', '%':
		return true
	}

	return false
}
