package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// chromeProfile is the profile these tests use. Every Chrome profile hands
// the extension the same bytes, so one covers them all. This one is the newest,
// so it is the wiring most likely to be wrong.
const chromeProfileName = "Chrome_152_PSK"

var chromeProfile = profiles.Chrome_152_PSK

// chromeRootStore39Capture is the payload stable Chrome 152.0.7977.64 on
// Android sent, from the "Unknown extension 51764" data field of a tls.peet.ws
// fingerprint. It carries Chrome Root Store 39 and no MTC anchors, so it holds
// the IDs of testdata/chrome_root_store_v39.pb.
//
// It is a historical fact and never changes. One capture is enough: another
// browser with the same store sends the same IDs in another order, and order is
// not what this checks.
const chromeRootStore39Capture = "00b80582df13020108839a648c9b2d010c08839a648c9b2d010704d679090c08839a648c9b2d010a04d679090b08839a648c9b2d010d0582df13020e08839a648c9b2d010b04d67909050582df13020d0582df13021404d679090404d679090804d679090d04d679090a04d679090708839a648c9b2d011204d67909010582df13020608839a648c9b2d01080582df13021208839a648c9b2d011304d679090f0582df13021308839a648c9b2d01090582df13020f04d6790906"

// chromeListPath is the list gen_chrome_trust_anchors.go writes.
const chromeListPath = "../trust_anchors/chrome_generated.go"

var chromeCaptureField = regexp.MustCompile(`chromeCapture = "([0-9a-f]+)"`)

var chromeVersionField = regexp.MustCompile(`// Chrome Root Store version: (\d+)`)

// chromeCapture reads the payload every Chrome profile must send from the
// generated list, so no copy of it lives in this file.
func chromeCapture(t *testing.T) string {
	t.Helper()

	source, err := os.ReadFile(chromeListPath)
	if err != nil {
		t.Fatal(err)
	}

	match := chromeCaptureField.FindSubmatch(source)
	if match == nil {
		t.Fatalf("%s holds no chromeCapture", chromeListPath)
	}

	return string(match[1])
}

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

// TestChromeTrustAnchors checks the trust_anchors extension (0xca34). The profile puts
// the IDs in a new order once per program run, the way Chromium's hash set order
// holds for the life of a process. Neither JA3 nor JA4 covers the contents of an
// extension, so client_test.go cannot detect a broken payload.
func TestChromeTrustAnchors(t *testing.T) {
	capture := chromeCapture(t)
	payloads := map[string]int{}

	for run := 1; run <= 2; run++ {
		t.Run(fmt.Sprintf("run_%d", run), func(t *testing.T) {
			payload := trustAnchorsFromSpec(t, chromeProfile)

			checkTrustAnchorIDs(t, chromeProfileName, capture, payload)

			payloads[hex.EncodeToString(payload)]++
		})
	}

	// One program run keeps one order, so every ClientHello of this test must
	// carry the same payload.
	if len(payloads) != 1 {
		t.Errorf("%s sent %d different payloads in one program run, expected 1", chromeProfileName, len(payloads))
	}

	// The IDs have far more orders than a test can meet by chance, so a payload
	// that repeats the capture means the reordering did not run.
	if payloads[capture] > 0 {
		t.Errorf("%s repeats the captured order, so the IDs were not reordered", chromeProfileName)
	}
}

// TestChromeTrustAnchorsOnTheWire checks what a server receives. It confirms that the
// payload survives the handshake and that the JA4 fingerprint does not move
// with it.
func TestChromeTrustAnchorsOnTheWire(t *testing.T) {
	capture := chromeCapture(t)

	var ja4Values []string

	for run := 1; run <= 2; run++ {
		t.Run(fmt.Sprintf("run_%d", run), func(t *testing.T) {
			options := []tls_client.HttpClientOption{
				skipPeetCertVerify,
				tls_client.WithClientProfile(chromeProfile),
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

			// tls.peet.ws reports the extension as "Unknown extension 51764"
			// today and may name it once the draft lands, so match the code
			// point in decimal and not the name around it.
			data := ""
			code := strconv.Itoa(trustAnchorsExtension)

			for _, extension := range tlsApiResponse.TLS.Extensions {
				if strings.Contains(extension.Name, code) {
					data = extension.Data
					break
				}
			}

			if data == "" {
				t.Fatalf("%s sent no trust_anchors extension", chromeProfileName)
			}

			payload, err := hex.DecodeString(data)
			if err != nil {
				t.Fatalf("%s sent %q, which is not hex", chromeProfileName, data)
			}

			checkTrustAnchorIDs(t, chromeProfileName, capture, payload)

			// The wire payload must be the one the profile holds.
			if want := hex.EncodeToString(trustAnchorsFromSpec(t, chromeProfile)); data != want {
				t.Errorf("%s sent %s, the profile holds %s", chromeProfileName, data, want)
			}

			if tlsApiResponse.TLS.Ja4 == "" {
				t.Fatalf("%s got no ja4 value from %s", chromeProfileName, peetApiEndpoint)
			}

			ja4Values = append(ja4Values, tlsApiResponse.TLS.Ja4)
		})
	}

	// JA4 counts extensions but not their contents, so the payload must not
	// move the fingerprint.
	for _, value := range ja4Values {
		if value != ja4Values[0] {
			t.Errorf("%s ja4 changed between connections: %v", chromeProfileName, ja4Values)
			break
		}
	}
}

// TestChromeTrustAnchorsAreAcceptedByAServer sends the extension to a server that reads
// it. Certificate verification stays on, so the whole handshake must hold.
//
// It covers the shape of the payload, not the IDs. A server that knows no ID in
// the list falls back to its normal certificate and still answers 204. The tests
// above cover the IDs.
func TestChromeTrustAnchorsAreAcceptedByAServer(t *testing.T) {
	client, err := tls_client.NewHttpClient(nil,
		tls_client.WithClientProfile(chromeProfile),
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
		t.Fatalf("%s could not reach %s: %s", chromeProfileName, generate204Endpoint, err)
	}

	defer resp.Body.Close()

	if _, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("%s could not read the body: %s", chromeProfileName, err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("%s got status %d from %s, expected %d", chromeProfileName, resp.StatusCode, generate204Endpoint, http.StatusNoContent)
	}
}

// TestChromeTrustAnchorsMatchUpstream runs the generator against the Chrome Root Store
// and compares its output with the committed list. It fails when Chromium
// changed the store, and when somebody changed the profiles without running the
// generator again.
//
// The failure is not a defect of this package. The message names the steps.
func TestChromeTrustAnchorsMatchUpstream(t *testing.T) {
	fresh := filepath.Join(t.TempDir(), "chrome_trust_anchors_generated.go")

	cmd := exec.Command("go", "run", "gen_chrome.go", "-out", fresh)
	cmd.Dir = "../trust_anchors"

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the generator failed: %s: %s", err, out)
	}

	want, err := os.ReadFile(chromeListPath)
	if err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(fresh)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) == string(want) {
		return
	}

	t.Errorf(`the Chrome Root Store is at version %s and no longer matches %s.

Do these steps:
1. Run "go generate ./trust_anchors".
2. Commit %s.
3. Name store version %s in the commit message.

Change nothing else. Do not touch testdata/chrome_root_store_v39.pb or
chromeRootStore39Capture. Those hold Chrome Root Store 39 and they must stay at
that version, because TestChromeTrustAnchorsSelectionRule needs a frozen input and a
frozen result.

Step 2 makes every Chrome profile advertise the new list.

committed:
%s

upstream:
%s
%s`,
		storeVersion(got), chromeListPath, chromeListPath, storeVersion(got),
		header(want), header(got), listDifference(t, want, got))
}

// listDifference names the IDs the two lists do not share. Two headers can read
// the same while the lists differ, so the message must name them.
func listDifference(t *testing.T, want, got []byte) string {
	t.Helper()

	ids := func(source []byte) map[string]bool {
		match := chromeCaptureField.FindSubmatch(source)
		if match == nil {
			return nil
		}

		payload, err := hex.DecodeString(string(match[1]))
		if err != nil {
			return nil
		}

		out := map[string]bool{}
		for _, id := range splitTrustAnchors(t, payload) {
			out[id] = true
		}

		return out
	}

	committed, upstream := ids(want), ids(got)

	var only []string

	for id := range committed {
		if !upstream[id] {
			only = append(only, "only in the committed list: "+id)
		}
	}

	for id := range upstream {
		if !committed[id] {
			only = append(only, "only upstream: "+id)
		}
	}

	if len(only) == 0 {
		return "The two lists hold the same IDs in another order."
	}

	sort.Strings(only)

	return strings.Join(only, "\n")
}

// storeVersion returns the Chrome Root Store version a generated list names.
func storeVersion(source []byte) string {
	match := chromeVersionField.FindSubmatch(source)
	if match == nil {
		return "unknown"
	}

	return string(match[1])
}

// header returns the part of the comment block that can move: the blob, the
// store version, and the anchor count. The "Code generated" and "Source" lines
// read the same in every list, so they stay out of the message.
func header(source []byte) string {
	var kept []string

	for _, line := range strings.Split(string(source), "\n") {
		if !strings.HasPrefix(line, "//") {
			break
		}

		if strings.HasPrefix(line, "// Code generated") || strings.HasPrefix(line, "// Source:") {
			continue
		}

		kept = append(kept, line)
	}

	return strings.Join(kept, "\n")
}

// TestChromeTrustAnchorsSelectionRule checks which fields the generator keeps:
// every trust_anchors entry, and an additional_certs entry only with
// tls_trust_anchor: true. Today 15 of the 28 IDs come from additional_certs.
//
// It runs the generator over a frozen crs.pb and compares the result with a
// browser capture of the same store. Every other test compares generated output
// with generated output. Input and result are frozen, so a root store update
// does not touch this test.
func TestChromeTrustAnchorsSelectionRule(t *testing.T) {
	fixture, err := filepath.Abs(filepath.Join("testdata", "chrome_root_store_v39.pb"))
	if err != nil {
		t.Fatal(err)
	}

	fresh := filepath.Join(t.TempDir(), "chrome_trust_anchors_generated.go")

	cmd := exec.Command("go", "run", "gen_chrome.go", "-source", fixture, "-out", fresh)
	cmd.Dir = "../trust_anchors"

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the generator failed: %s: %s", err, out)
	}

	source, err := os.ReadFile(fresh)
	if err != nil {
		t.Fatal(err)
	}

	match := chromeCaptureField.FindSubmatch(source)
	if match == nil {
		t.Fatalf("the generator wrote no chromeCapture for %s", fixture)
	}

	generated, err := hex.DecodeString(string(match[1]))
	if err != nil {
		t.Fatal(err)
	}

	captured, err := hex.DecodeString(chromeRootStore39Capture)
	if err != nil {
		t.Fatal(err)
	}

	got := sortedTrustAnchorIDs(t, generated)
	want := sortedTrustAnchorIDs(t, captured)

	if len(got) != len(want) {
		t.Fatalf("the frozen root store gives %d IDs, the browser sent %d", len(got), len(want))
	}

	for i, id := range got {
		if id != want[i] {
			t.Fatalf("the frozen root store gives %v, the browser sent %v", got, want)
		}
	}
}
