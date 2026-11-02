package tls_client

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
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

// generateH2GreaseSettingID generates a valid 16-bit GREASE setting ID for HTTP/2.
// Per RFC 8701, the reserved 16-bit GREASE values are 0x0a0a, 0x1a1a, ..., 0xfafa
// (both bytes equal, each of the form 0xN0x0a where N is 0..15). Real Chrome
// randomly picks one of these per connection for its HTTP/2 SETTINGS frame.
// HTTP/2 setting IDs are 16-bit, so this differs from the HTTP/3 path which uses
// large varint GREASE IDs (see generateGREASESettingID).
func generateH2GreaseSettingID() uint16 {
	nBig, _ := rand.Int(rand.Reader, big.NewInt(16))
	b := uint8(0x0a) + uint8(nBig.Uint64())*0x10
	return uint16(b)<<8 | uint16(b)
}
