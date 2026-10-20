package tests

import (
	"net"
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"
	tls "github.com/bogdanfinn/utls"
)

// greaseSignatureSchemeProfiles are the profiles that send a GREASE value as
// their first signature algorithm. Add a row for a new browser version.
var greaseSignatureSchemeProfiles = map[string]profiles.ClientProfile{
	"Chrome_152":     profiles.Chrome_152,
	"Chrome_152_PSK": profiles.Chrome_152_PSK,
}

// TestGreaseSignatureSchemeIsRandom checks the first entry of
// signature_algorithms. Chrome sends a GREASE value there and picks a new one
// for every connection.
//
// A spec holds the GREASE placeholder. utls draws the value in ApplyPreset,
// from the seed of the connection, so the test applies every spec to a UConn
// over a pipe. No byte reaches the pipe, because ApplyPreset performs no IO.
//
// TestGreaseSignatureAlgorithmOnTheWire checks the same value over the wire,
// but it can only afford a handful of connections. This test draws enough
// values to show that all 16 of them appear.
func TestGreaseSignatureSchemeIsRandom(t *testing.T) {
	for name, profile := range greaseSignatureSchemeProfiles {
		t.Run(name, func(t *testing.T) {
			greaseSignatureSchemeIsRandom(t, name, profile)
		})
	}
}

func greaseSignatureSchemeIsRandom(t *testing.T, name string, profile profiles.ClientProfile) {
	t.Helper()

	seen := map[uint64]int{}

	for i := 0; i < 2000; i++ {
		spec, err := profile.GetClientHelloSpec()
		if err != nil {
			t.Fatal(err)
		}

		value := uint64(firstSignatureAlgorithm(t, name, spec))

		if !isGreaseValue(value) {
			t.Fatalf("%s first signature algorithm is 0x%04x, expected a GREASE value", name, value)
		}

		seen[value]++
	}

	// There are 16 GREASE values. Over 2000 draws every one of them should
	// appear, otherwise the value is not being randomized.
	if len(seen) != 16 {
		t.Errorf("%s used %d of the 16 GREASE values: %v", name, len(seen), seen)
	}
}

// firstSignatureAlgorithm applies a spec to a connection and returns the first
// signature algorithm the ClientHello would carry.
func firstSignatureAlgorithm(t *testing.T, name string, spec tls.ClientHelloSpec) tls.SignatureScheme {
	t.Helper()

	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	uconn := tls.UClient(client, &tls.Config{ServerName: "example.com"}, tls.HelloCustom, false, false, false)

	if err := uconn.ApplyPreset(&spec); err != nil {
		t.Fatal(err)
	}

	for _, extension := range uconn.Extensions {
		if signatureAlgorithms, ok := extension.(*tls.SignatureAlgorithmsExtension); ok {
			if len(signatureAlgorithms.SupportedSignatureAlgorithms) == 0 {
				t.Fatalf("%s sends an empty signature_algorithms extension", name)
			}

			return signatureAlgorithms.SupportedSignatureAlgorithms[0]
		}
	}

	t.Fatalf("%s sends no signature_algorithms extension", name)

	return 0
}
