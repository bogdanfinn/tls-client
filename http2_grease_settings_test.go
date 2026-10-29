package tls_client

import (
	"testing"

	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
)

func TestChromiumHTTP2SettingsAddsGrease(t *testing.T) {
	settings := map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:   65536,
		http2.SettingEnablePush:        0,
		http2.SettingInitialWindowSize: 6291456,
		http2.SettingMaxHeaderListSize: 262144,
	}
	order := []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	}

	for _, client := range []string{chromeClientName, "Edge", "Opera"} {
		gotSettings, gotOrder := chromiumHTTP2Settings(tls.ClientHelloID{Client: client}, settings, order)

		if len(gotSettings) != len(settings)+1 {
			t.Fatalf("client %q: expected %d settings, got %d", client, len(settings)+1, len(gotSettings))
		}

		if len(gotOrder) != len(order)+1 {
			t.Fatalf("client %q: expected %d ordered settings, got %d", client, len(order)+1, len(gotOrder))
		}

		// The GREASE entry sits last, where Chromium puts it.
		greaseID := gotOrder[len(gotOrder)-1]
		if uint16(greaseID)&0x0f0f != 0x0a0a {
			t.Fatalf("client %q: last setting 0x%04x is not a GREASE setting ID", client, uint16(greaseID))
		}

		value, ok := gotSettings[greaseID]
		if !ok {
			t.Fatalf("client %q: GREASE setting 0x%04x is in the order but not in the map", client, uint16(greaseID))
		}
		if value == 0 {
			t.Fatalf("client %q: GREASE setting value must not be 0", client)
		}

		// The original settings keep their values.
		for _, id := range order {
			if gotSettings[id] != settings[id] {
				t.Fatalf("client %q: setting 0x%04x changed value", client, uint16(id))
			}
		}
	}

	// The shared profile inputs are not mutated.
	if len(settings) != 4 || len(order) != 4 {
		t.Fatal("input settings or order were mutated")
	}
}

func TestChromiumHTTP2SettingsLeavesOtherClientsAlone(t *testing.T) {
	settings := map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:   65536,
		http2.SettingInitialWindowSize: 6291456,
	}
	order := []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingInitialWindowSize,
	}

	for _, client := range []string{firefoxClientName, "Safari", "iOS", "Android", "MMSIos", "OkHttp4Android13", "Custom"} {
		gotSettings, gotOrder := chromiumHTTP2Settings(tls.ClientHelloID{Client: client}, settings, order)

		if len(gotSettings) != len(settings) || len(gotOrder) != len(order) {
			t.Fatalf("client %q got a GREASE setting it should not send", client)
		}
	}
}

func TestChromiumHTTP2SettingsWithoutSettings(t *testing.T) {
	gotSettings, gotOrder := chromiumHTTP2Settings(tls.ClientHelloID{Client: chromeClientName}, nil, nil)

	if len(gotSettings) != 0 || len(gotOrder) != 0 {
		t.Fatal("empty settings should stay empty")
	}
}
