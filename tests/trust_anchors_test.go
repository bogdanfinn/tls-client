package tests

import (
	"bytes"
	"encoding/hex"
	"slices"
	"testing"

	"github.com/bogdanfinn/tls-client/trust_anchors"
)

// TestSort checks that the IDs decide the order, not the records
// that carry them. Chrome 154 and later sort the IDs.
func TestSort(t *testing.T) {
	// ID 03 is one byte, ID 0299 is two. The record of 03 is 0103 and the record
	// of 0299 is 020299, so the records and the IDs sort in opposite orders.
	records, err := trust_anchors.Split("00050103020299")
	if err != nil {
		t.Fatal(err)
	}

	got := hex.EncodeToString(trust_anchors.Sort(records))

	if want := "00050202990103"; got != want {
		t.Errorf("trust_anchors.Sort gave %s, expected %s", got, want)
	}

	// A second call gives the same payload, unlike trust_anchors.Shuffle.
	if again := hex.EncodeToString(trust_anchors.Sort(records)); again != got {
		t.Errorf("two calls gave %s and %s", got, again)
	}
}

// TestSortKeepsEveryID checks the Chrome list itself.
func TestSortKeepsEveryID(t *testing.T) {
	records, err := trust_anchors.Split(chromeCapture(t))
	if err != nil {
		t.Fatal(err)
	}

	sorted, err := trust_anchors.Split(hex.EncodeToString(trust_anchors.Sort(records)))
	if err != nil {
		t.Fatal(err)
	}

	if len(sorted) != len(records) {
		t.Fatalf("the list holds %d records, the sorted payload holds %d", len(records), len(sorted))
	}

	for i := 1; i < len(sorted); i++ {
		if bytes.Compare(sorted[i-1][1:], sorted[i][1:]) >= 0 {
			t.Fatalf("record %d does not come after record %d", i, i-1)
		}
	}
}

// TestChromeSortedHoldsTheSameAnchors checks that ChromeSorted and
// ChromeShuffled carry one list in two orders.
func TestChromeSortedHoldsTheSameAnchors(t *testing.T) {
	sorted, err := trust_anchors.Split(hex.EncodeToString(trust_anchors.ChromeSorted()))
	if err != nil {
		t.Fatal(err)
	}

	shuffled, err := trust_anchors.Split(hex.EncodeToString(trust_anchors.ChromeShuffled()))
	if err != nil {
		t.Fatal(err)
	}

	if len(sorted) != len(shuffled) {
		t.Fatalf("ChromeSorted holds %d anchors, ChromeShuffled holds %d", len(sorted), len(shuffled))
	}

	for i := 1; i < len(sorted); i++ {
		if bytes.Compare(sorted[i-1][1:], sorted[i][1:]) >= 0 {
			t.Fatalf("ChromeSorted record %d does not come after record %d", i, i-1)
		}
	}

	slices.SortFunc(shuffled, func(a, b []byte) int { return bytes.Compare(a[1:], b[1:]) })

	for i := range sorted {
		if !bytes.Equal(sorted[i], shuffled[i]) {
			t.Fatalf("anchor %d is %x in ChromeSorted and %x in ChromeShuffled", i, sorted[i], shuffled[i])
		}
	}
}
