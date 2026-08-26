package tests

import (
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

// TestGreaseSignatureSchemeIsRandom builds ClientHello specs offline and checks
// the first entry of signature_algorithms. Chrome sends a GREASE value there
// and picks a new one for every connection.
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

		var schemes []tls.SignatureScheme

		for _, extension := range spec.Extensions {
			if signatureAlgorithms, ok := extension.(*tls.SignatureAlgorithmsExtension); ok {
				schemes = signatureAlgorithms.SupportedSignatureAlgorithms
				break
			}
		}

		if len(schemes) == 0 {
			t.Fatalf("%s sends no signature_algorithms extension", name)
		}

		value := uint64(schemes[0])

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
