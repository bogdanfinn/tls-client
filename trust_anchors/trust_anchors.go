package trust_anchors

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"slices"
)

//go:generate go run gen_chrome.go

// The payload of the trust_anchors extension (0xca34) is a
// RequestedTrustAnchorList: a 16-bit list length, then an 8-bit length and an
// ID per anchor. The format is the same for every browser. This list is the
// Chrome one, and gen_chrome.go writes it from the Chrome Root Store.
//
// The name says the order. Chrome 153 and earlier draw a new order per process,
// Chrome 154 and later sort. See Shuffle and Sort.
// https://source.chromium.org/search?q=TLSEXT_TYPE_trust_anchors
// https://issues.chromium.org/issues/398275713
var chromeShuffled = Shuffle(mustSplit(chromeCapture))

// The same list in the Chrome 154 order. See Sort.
var chromeSorted = Sort(mustSplit(chromeCapture))

// ChromeShuffled returns the payload a Chrome 153 or earlier profile sends.
// Every call gives a copy, so a caller cannot change what the next call returns.
func ChromeShuffled() []byte {
	return bytes.Clone(chromeShuffled)
}

// ChromeSorted returns the payload a Chrome 154 or later profile sends. Every
// call gives a copy, and every process gives the same bytes.
func ChromeSorted() []byte {
	return bytes.Clone(chromeSorted)
}

// mustSplit splits a payload into records. It panics on malformed
// input, which can only come from a literal in this package.
func mustSplit(capture string) [][]byte {
	records, err := Split(capture)
	if err != nil {
		panic(err)
	}

	return records
}

// Split splits a payload into records. A record is the 8-bit length
// with the ID that follows it, so records reorder without further parsing.
func Split(capture string) ([][]byte, error) {
	payload, err := hex.DecodeString(capture)
	if err != nil {
		return nil, fmt.Errorf("trust anchors payload is not valid hex: %w", err)
	}

	if len(payload) < 2 {
		return nil, errors.New("trust anchors payload is shorter than its length field")
	}

	list := payload[2:]
	if int(payload[0])<<8|int(payload[1]) != len(list) {
		return nil, errors.New("trust anchors length field does not match the list")
	}

	var records [][]byte

	for i := 0; i < len(list); {
		end := i + 1 + int(list[i])
		if end > len(list) {
			return nil, errors.New("trust anchor ID runs past the end of the list")
		}

		records = append(records, list[i:end])
		i = end
	}

	return records, nil
}

// BuildPayload turns a capture into extension data. The capture is the hex
// string from the "Unknown extension 51764" data field of a browser
// fingerprint, from the 16-bit list length on.
//
// A sorted capture comes from Chrome 154 or later and keeps its order. Any
// other capture is shuffled. One process keeps one order, so call this once and
// reuse the result.
// Order alone tells the two apart.
func BuildPayload(capture string) ([]byte, error) {
	records, err := Split(capture)
	if err != nil {
		return nil, err
	}

	if slices.IsSortedFunc(records, byID) {
		return Encode(records), nil
	}

	return Shuffle(records), nil
}

// byID orders two records by the anchor ID after the length byte. Chromium
// sorts the IDs, not the records that carry them, so a short ID does not come
// first only because its length byte is smaller. BuildPayload detects the order
// with this function, and Sort makes it, so the two agree.
func byID(a, b []byte) int {
	return bytes.Compare(a[1:], b[1:])
}

// Encode builds the payload from the records in the order they hold.
func Encode(records [][]byte) []byte {
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

// Sort encodes the records with the IDs in ascending order. Chrome
// 154 and later sort the list before they encode the extension, so one set of
// IDs gives one payload and it does not change between processes.
// https://chromium.googlesource.com/chromium/src/+/942bda4298c165efb3635f4149cdbc31a259e6b2
func Sort(anchors [][]byte) []byte {
	records := make([][]byte, len(anchors))
	copy(records, anchors)

	slices.SortFunc(records, byID)

	return Encode(records)
}

// Shuffle encodes the records in a new order. Chrome 153 and earlier
// write the IDs in absl::flat_hash_set iteration order, which holds for the life
// of a process and differs between processes, so the callers shuffle once at
// package load.
//
// Chrome 154 and later sort the IDs instead. A profile for one of those versions
// must call Sort.
//
// It panics if the random source fails, which crypto/rand does not survive
// either.
func Shuffle(anchors [][]byte) []byte {
	records := make([][]byte, len(anchors))
	copy(records, anchors)

	for i := len(records) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			panic(err)
		}

		records[i], records[j.Int64()] = records[j.Int64()], records[i]
	}

	return Encode(records)
}
