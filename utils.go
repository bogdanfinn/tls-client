package tls_client

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"

	"github.com/bogdanfinn/fhttp/http2"
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

// generateGREASEH2SettingID returns a random 16-bit GREASE setting ID for the
// HTTP/2 SETTINGS frame. RFC 8701 reserves the IDs 0x?a?a, one nibble free, so
// the result is one of 0x0a0a through 0xfafa. It panics if the random source
// fails, which crypto/rand does not survive either.
func generateGREASEH2SettingID() http2.SettingID {
	n, err := rand.Int(rand.Reader, big.NewInt(16))
	if err != nil {
		panic(err)
	}

	return http2.SettingID(n.Uint64()<<12 | n.Uint64()<<4 | 0x0a0a)
}
