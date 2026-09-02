// Package main shows how to build a client profile from a raw JA3 string
// instead of picking one of the bundled browser profiles. Use this when you
// need a fingerprint that isn't shipped with the library.
//
// Run: go run ./example/custom_profile
package main

import (
	"io"
	"log"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	tls "github.com/bogdanfinn/utls"
)

// A JA3 string names extension IDs but carries no extension data. 51764 is
// trust_anchors, which Chrome 144 and later send, so its payload has to be
// passed alongside the JA3 string - see trustAnchorsPayload below.
//
// GREASE values are written as 2570 and may appear more than once in the
// extension list; utls draws a real value per connection for each of them.
// If you need an extension whose data a JA3 string cannot express - PSK or
// ALPS, for instance - build the spec by hand instead: ./example/custom_spec.
const ja3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-10-11-13-16-23-43-51-65281-45-21-51764,29-23-24,0"

// The data field of the "Unknown extension 51764" entry of a browser
// fingerprint, starting at the 16-bit list length. "0000" is an empty anchor
// list. Leave it empty when your JA3 string does not list 51764.
const trustAnchorsPayload = "001904d67909010582df13020604d679090a08839a648c9b2d0107"

func main() {
	specFactory, err := tls_client.GetSpecFactoryFromJa3String(
		ja3,
		// Chrome 150 and later advertise the ML-DSA codepoints, and Chrome 152
		// puts a GREASE value first. utls draws a fresh GREASE value per
		// connection, just like it does for the bundled profiles.
		[]string{"GREASE", "MLDSA65", "ECDSAWithP256AndSHA256", "PSSWithSHA256", "PKCS1WithSHA256"}, // supportedSignatureAlgorithms
		nil,                              // supportedDelegatedCredentialsAlgorithms
		[]string{"GREASE", "1.3", "1.2"}, // supportedVersions
		[]string{"GREASE", "X25519"},     // keyShareCurves
		[]string{"h2", "http/1.1"},       // alpnProtocols
		[]string{"h2"},                   // alpsProtocols
		nil,                              // echCandidateCipherSuites
		nil,                              // echCandidatePayloads
		[]string{"zlib"},                 // certCompressionAlgorithms
		0,                                // recordSizeLimit
		trustAnchorsPayload,              // trustAnchorsPayload, required when the JA3 string lists 51764
	)
	if err != nil {
		log.Fatal(err)
	}

	clientHelloID := tls.ClientHelloID{
		Client:      "MyCustomProfile",
		Version:     "1",
		SpecFactory: specFactory,
	}

	settings := map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	}
	settingsOrder := []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	}
	pseudoHeaderOrder := []string{":method", ":authority", ":scheme", ":path"}

	// The trailing arguments cover the HTTP/3 side of the profile
	// (h3 settings, their order, the priority param, the h3 pseudo header
	// order and whether to send GREASE frames). Passing zero values means the
	// profile has no HTTP/3 fingerprint, which is fine as long as you do not
	// use HTTP/3 - see ./example/protocol_racing.
	customProfile := profiles.NewClientProfile(clientHelloID, settings, settingsOrder, pseudoHeaderOrder, 15663105, nil, nil, 0, false, nil, nil, 0, nil, false)

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(customProfile),
	}...)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("status: %d\n%s\n", resp.StatusCode, body)
}
