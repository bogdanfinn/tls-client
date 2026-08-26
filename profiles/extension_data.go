package profiles

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math/big"
)

// Captured payloads for the trust_anchors extension (0xca34). A payload is a
// RequestedTrustAnchorList: a 16-bit list length, then an 8-bit length and an
// ID per anchor.
// https://source.chromium.org/search?q=TLSEXT_TYPE_trust_anchors
// https://issues.chromium.org/issues/398275713
//
// To add a capture for a new browser version:
//
//  1. Open https://tls.peet.ws/api/all in that browser, or in a local container
//     that runs the same service.
//  2. In the tls.extensions array, find the entry named "Unknown extension
//     51764" and copy its data field. The field is the payload in hex and it
//     starts at the 16-bit list length, so copy it whole.
//  3. Add a const with the hex string and a payload var next to the entries
//     below. The var fails at package load if the payload is malformed, so a
//     wrong copy shows at once.
//  4. Pass the payload var as the extension data in the profile.

// Stable Chrome 152.0.7977.64 Android. 186 bytes, 28 IDs, Chrome Root Store 39,
// no MTC anchors.
const chrome152TrustAnchorsCapture = "00b80582df13020108839a648c9b2d010c08839a648c9b2d010704d679090c08839a648c9b2d010a04d679090b08839a648c9b2d010d0582df13020e08839a648c9b2d010b04d67909050582df13020d0582df13021404d679090404d679090804d679090d04d679090a04d679090708839a648c9b2d011204d67909010582df13020608839a648c9b2d01080582df13021208839a648c9b2d011304d679090f0582df13021308839a648c9b2d01090582df13020f04d6790906"

var chrome152TrustAnchors = shuffleTrustAnchors(mustSplitTrustAnchors(chrome152TrustAnchorsCapture))

// mustDecodeHex converts a wire-format hex string into extension data. It
// panics on malformed input, which can only come from a literal in this
// package.
func mustDecodeHex(s string) []byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}

	return data
}

// mustSplitTrustAnchors splits a captured payload into records. A record is the
// 8-bit length with the ID that follows it, so records reorder without further
// parsing. It panics on malformed input, which can only come from a literal in
// this package.
func mustSplitTrustAnchors(capture string) [][]byte {
	payload := mustDecodeHex(capture)
	if len(payload) < 2 {
		panic("trust anchors payload is shorter than its length field")
	}

	list := payload[2:]
	if int(payload[0])<<8|int(payload[1]) != len(list) {
		panic("trust anchors length field does not match the list")
	}

	var records [][]byte

	for i := 0; i < len(list); {
		end := i + 1 + int(list[i])
		if end > len(list) {
			panic("trust anchor ID runs past the end of the list")
		}

		records = append(records, list[i:end])
		i = end
	}

	return records
}

// shuffleTrustAnchors builds the trust_anchors payload from the given records
// and gives them a new order. Chromium writes the IDs in absl::flat_hash_set
// iteration order, which holds for the life of the process and differs between
// processes. The callers above therefore shuffle once at package load, so one
// program run keeps one order and two runs differ.
//
// It panics if the random source fails, which crypto/rand does not survive
// either.
func shuffleTrustAnchors(anchors [][]byte) []byte {
	records := make([][]byte, len(anchors))
	copy(records, anchors)

	for i := len(records) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			panic(err)
		}

		records[i], records[j.Int64()] = records[j.Int64()], records[i]
	}

	size := 0
	for _, record := range records {
		size += len(record)
	}

	payload := make([]byte, 2, 2+size)
	binary.BigEndian.PutUint16(payload, uint16(size))

	for _, record := range records {
		payload = append(payload, record...)
	}

	return payload
}
