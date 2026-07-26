package tls_client

import (
	"testing"

	"github.com/bogdanfinn/fhttp/http2"
)

// rfc8701H2GreaseIDs are the reserved 16-bit GREASE values per RFC 8701.
var rfc8701H2GreaseIDs = map[http2.SettingID]bool{
	0x0a0a: true, 0x1a1a: true, 0x2a2a: true, 0x3a3a: true,
	0x4a4a: true, 0x5a5a: true, 0x6a6a: true, 0x7a7a: true,
	0x8a8a: true, 0x9a9a: true, 0xaaaa: true, 0xbaba: true,
	0xcaca: true, 0xdada: true, 0xeaea: true, 0xfafa: true,
}

func TestApplyH2GreaseSetting_Chrome(t *testing.T) {
	base := map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:   65536,
		http2.SettingEnablePush:        0,
		http2.SettingInitialWindowSize: 6291456,
		http2.SettingMaxHeaderListSize: 262144,
	}
	baseOrder := []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	}

	got, gotOrder := applyH2GreaseSetting(base, baseOrder, true)

	// All original settings preserved.
	for k, v := range base {
		if got[k] != v {
			t.Errorf("setting %d changed: got %d, want %d", k, got[k], v)
		}
	}

	// Exactly one GREASE entry appended at the end.
	if len(gotOrder) != len(baseOrder)+1 {
		t.Fatalf("order length: got %d, want %d", len(gotOrder), len(baseOrder)+1)
	}
	greaseID := gotOrder[len(baseOrder)]
	if !rfc8701H2GreaseIDs[greaseID] {
		t.Errorf("appended setting ID 0x%04x is not a valid RFC 8701 GREASE value", greaseID)
	}
	greaseVal, ok := got[greaseID]
	if !ok {
		t.Fatalf("GREASE setting %d not present in settings map", greaseID)
	}
	if greaseVal == 0 {
		t.Errorf("GREASE setting value must be non-zero, got 0")
	}

	// Original settings order preserved as a prefix.
	for i, id := range baseOrder {
		if gotOrder[i] != id {
			t.Errorf("order[%d]: got %d, want %d", i, gotOrder[i], id)
		}
	}

	// Original map and slice must not be mutated.
	if _, ok := base[greaseID]; ok {
		t.Errorf("original settings map was mutated with GREASE entry")
	}
	if len(baseOrder) != 4 {
		t.Errorf("original settingsOrder was mutated: got len %d, want 4", len(baseOrder))
	}
}

func TestApplyH2GreaseSetting_NonChrome(t *testing.T) {
	base := map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:   65536,
		http2.SettingEnablePush:        0,
		http2.SettingInitialWindowSize: 131072,
		http2.SettingMaxFrameSize:      16384,
	}
	baseOrder := []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingInitialWindowSize,
		http2.SettingMaxFrameSize,
	}

	got, gotOrder := applyH2GreaseSetting(base, baseOrder, false)

	// Non-Chrome profiles must receive no GREASE: returned unchanged.
	if len(got) != len(base) {
		t.Errorf("settings map size changed: got %d, want %d", len(got), len(base))
	}
	if len(gotOrder) != len(baseOrder) {
		t.Errorf("order length changed: got %d, want %d", len(gotOrder), len(baseOrder))
	}
	for _, id := range gotOrder {
		if rfc8701H2GreaseIDs[id] {
			t.Errorf("non-Chrome profile received GREASE setting 0x%04x", id)
		}
	}
}
