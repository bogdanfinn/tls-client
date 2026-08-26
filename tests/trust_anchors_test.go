package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	tls "github.com/bogdanfinn/utls"
)

// trustAnchorsExtension is the code point of the trust_anchors extension.
const trustAnchorsExtension = 0xca34

// generate204Endpoint answers with 204 and an empty body. Google runs the TLS
// trust anchor IDs draft, so it reads the extension instead of skipping it.
const generate204Endpoint = "https://www.google.com/generate_204"

// trustAnchorProfiles are the profiles that send the trust_anchors extension.
// Add a row for a new browser version, together with its capture below.
var trustAnchorProfiles = []struct {
	name    string
	profile profiles.ClientProfile
	capture string
}{
	{"Chrome_152", profiles.Chrome_152, chrome152TrustAnchorsCapture},
	{"Chrome_152_PSK", profiles.Chrome_152_PSK, chrome152TrustAnchorsCapture},
}

// chrome152TrustAnchorsCapture is the payload as stable Chrome 152.0.7977.64 on
// Android sent it. The profile keeps its IDs and gives them a new order, so a
// ClientHello that repeats this exact order means the reordering did not run.
const chrome152TrustAnchorsCapture = "00b80582df13020108839a648c9b2d010c08839a648c9b2d010704d679090c08839a648c9b2d010a04d679090b08839a648c9b2d010d0582df13020e08839a648c9b2d010b04d67909050582df13020d0582df13021404d679090404d679090804d679090d04d679090a04d679090708839a648c9b2d011204d67909010582df13020608839a648c9b2d01080582df13021208839a648c9b2d011304d679090f0582df13021308839a648c9b2d01090582df13020f04d6790906"

// splitTrustAnchors parses a RequestedTrustAnchorList the way a server does. It
// returns the IDs in the order they arrived and fails the test on a malformed
// payload.
func splitTrustAnchors(t *testing.T, payload []byte) []string {
	t.Helper()

	if len(payload) < 2 {
		t.Fatalf("payload of %d bytes is shorter than its length field", len(payload))
	}

	list := payload[2:]
	if length := int(payload[0])<<8 | int(payload[1]); length != len(list) {
		t.Fatalf("length field says %d bytes, list holds %d", length, len(list))
	}

	var ids []string

	for i := 0; i < len(list); {
		end := i + 1 + int(list[i])
		if end > len(list) {
			t.Fatalf("ID at offset %d runs past the end of the list", i)
		}

		ids = append(ids, hex.EncodeToString(list[i+1:end]))
		i = end
	}

	return ids
}

// checkTrustAnchorIDs fails the test unless the payload holds the IDs of the
// capture, each of them once and none of them missing. The order is free,
// because Chromium writes the IDs in hash set iteration order.
func checkTrustAnchorIDs(t *testing.T, name string, capture string, payload []byte) {
	t.Helper()

	want, err := hex.DecodeString(capture)
	if err != nil {
		t.Fatal(err)
	}

	got := sortedTrustAnchorIDs(t, payload)
	expected := sortedTrustAnchorIDs(t, want)

	if len(got) != len(expected) {
		t.Fatalf("%s sent %d trust anchor IDs, expected %d", name, len(got), len(expected))
	}

	for i, id := range got {
		if id != expected[i] {
			t.Errorf("%s sent %v, expected %v", name, got, expected)
			return
		}
	}
}

// sortedTrustAnchorIDs parses a payload and returns its IDs in sorted order.
func sortedTrustAnchorIDs(t *testing.T, payload []byte) []string {
	t.Helper()

	ids := splitTrustAnchors(t, payload)
	sort.Strings(ids)

	return ids
}

// trustAnchorsFromSpec returns the payload of the single trust_anchors
// extension of the ClientHello.
func trustAnchorsFromSpec(t *testing.T, profile profiles.ClientProfile) []byte {
	t.Helper()

	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		t.Fatal(err)
	}

	var payloads [][]byte

	for _, extension := range spec.Extensions {
		if generic, ok := extension.(*tls.GenericExtension); ok && generic.Id == trustAnchorsExtension {
			payloads = append(payloads, generic.Data)
		}
	}

	if len(payloads) != 1 {
		t.Fatalf("%s sends %d trust_anchors extensions, expected 1", profile.GetClientHelloStr(), len(payloads))
	}

	return payloads[0]
}

// TestTrustAnchors checks the trust_anchors extension (0xca34). A profile holds
// one captured payload and puts its IDs in a new order once per program run,
// the way Chromium's hash set iteration order holds for the life of a process.
// Neither JA3 nor JA4 covers the contents of an extension, so the fingerprint
// assertions in client_test.go cannot detect a broken payload.
func TestTrustAnchors(t *testing.T) {
	for _, tc := range trustAnchorProfiles {
		payloads := map[string]int{}

		for run := 1; run <= 5; run++ {
			t.Run(fmt.Sprintf("%s/run_%d", tc.name, run), func(t *testing.T) {
				payload := trustAnchorsFromSpec(t, tc.profile)

				checkTrustAnchorIDs(t, tc.name, tc.capture, payload)

				payloads[hex.EncodeToString(payload)]++
			})
		}

		// One program run keeps one order, so every ClientHello of this test
		// must carry the same payload.
		if len(payloads) != 1 {
			t.Errorf("%s sent %d different payloads in one program run, expected 1", tc.name, len(payloads))
		}

		// The IDs have far more orders than a test can meet by chance, so a
		// payload that repeats the capture means the reordering did not run.
		if payloads[tc.capture] > 0 {
			t.Errorf("%s repeats the captured order, so the IDs were not reordered", tc.name)
		}
	}
}

// TestTrustAnchorsOnTheWire checks what a server receives. It confirms that the
// payload survives the handshake and that the JA4 fingerprint does not move
// with it.
func TestTrustAnchorsOnTheWire(t *testing.T) {
	ja4Values := map[string][]string{}
	payloads := map[string]int{}

	for _, tc := range trustAnchorProfiles {
		for run := 1; run <= 2; run++ {
			t.Run(fmt.Sprintf("%s/run_%d", tc.name, run), func(t *testing.T) {
				options := []tls_client.HttpClientOption{
					skipPeetCertVerify,
					tls_client.WithClientProfile(tc.profile),
					tls_client.WithTimeoutSeconds(120),
				}

				client, err := tls_client.NewHttpClient(nil, options...)
				if err != nil {
					t.Fatal(err)
				}

				req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
				if err != nil {
					t.Fatal(err)
				}

				req.Header = defaultHeader

				resp, err := client.Do(req)
				if err != nil {
					t.Fatal(err)
				}

				defer resp.Body.Close()

				readBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}

				tlsApiResponse := TlsApiResponse{}
				if err := json.Unmarshal(readBytes, &tlsApiResponse); err != nil {
					t.Fatal(err)
				}

				// tls.peet.ws reports the extension as "Unknown extension
				// 51764" today and may name it once the draft lands, so match
				// the code point in decimal and not the name around it.
				data := ""
				code := strconv.Itoa(trustAnchorsExtension)

				for _, extension := range tlsApiResponse.TLS.Extensions {
					if strings.Contains(extension.Name, code) {
						data = extension.Data
						break
					}
				}

				if data == "" {
					t.Fatalf("%s sent no trust_anchors extension", tc.name)
				}

				payload, err := hex.DecodeString(data)
				if err != nil {
					t.Fatalf("%s sent %q, which is not hex", tc.name, data)
				}

				checkTrustAnchorIDs(t, tc.name, tc.capture, payload)

				// The wire payload must be the one the profile holds.
				if want := hex.EncodeToString(trustAnchorsFromSpec(t, tc.profile)); data != want {
					t.Errorf("%s sent %s, the profile holds %s", tc.name, data, want)
				}

				payloads[tc.name+" "+data]++

				if tlsApiResponse.TLS.Ja4 == "" {
					t.Fatalf("%s got no ja4 value from %s", tc.name, peetApiEndpoint)
				}

				ja4Values[tc.name] = append(ja4Values[tc.name], tlsApiResponse.TLS.Ja4)
			})
		}
	}

	if len(payloads) != len(trustAnchorProfiles) {
		t.Errorf("one program run sent %d payloads, expected one per profile", len(payloads))
	}

	// JA4 counts extensions but not their contents, so the payload must not
	// move the fingerprint.
	for name, values := range ja4Values {
		for _, value := range values {
			if value != values[0] {
				t.Errorf("%s ja4 changed between connections: %v", name, values)
				break
			}
		}
	}
}

// TestTrustAnchorsAreAcceptedByAServer sends the extension to a server that
// reads it and checks that the handshake and the request go through.
// Certificate verification stays on, so the whole handshake must hold.
//
// The test covers the shape of the payload, not the IDs in it. A server that
// finds no ID it knows falls back to its normal certificate, so a list of
// unknown IDs still answers with 204. The tests above cover the IDs.
func TestTrustAnchorsAreAcceptedByAServer(t *testing.T) {
	for _, tc := range trustAnchorProfiles {
		t.Run(tc.name, func(t *testing.T) {
			client, err := tls_client.NewHttpClient(nil,
				tls_client.WithClientProfile(tc.profile),
				tls_client.WithTimeoutSeconds(30),
			)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest(http.MethodGet, generate204Endpoint, nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header = defaultHeader

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("%s could not reach %s: %s", tc.name, generate204Endpoint, err)
			}

			defer resp.Body.Close()

			if _, err := io.ReadAll(resp.Body); err != nil {
				t.Fatalf("%s could not read the body: %s", tc.name, err)
			}

			if resp.StatusCode != http.StatusNoContent {
				t.Errorf("%s got status %d from %s, expected %d", tc.name, resp.StatusCode, generate204Endpoint, http.StatusNoContent)
			}
		})
	}
}
