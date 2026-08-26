package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

// isGreaseValue reports whether v is one of the 16 GREASE values defined in
// RFC 8701: both bytes are equal and the low nibble is 0xa.
func isGreaseValue(v uint64) bool {
	return v>>8 == v&0xff && v&0xf == 0xa
}

// TestGreaseSignatureAlgorithmOnTheWire checks that the profile sends a GREASE
// value as the first entry of the signature_algorithms extension, and that the
// value changes between connections without moving the JA4 fingerprint. Chrome
// 152 added that value. Neither JA3 nor JA4 covers the contents of
// signature_algorithms, so the fingerprint assertions in client_test.go cannot
// detect it.
func TestGreaseSignatureAlgorithmOnTheWire(t *testing.T) {
	ja4Values := map[string][]string{}

	for name, profile := range greaseSignatureSchemeProfiles {
		for run := 1; run <= 3; run++ {
			t.Run(fmt.Sprintf("%s/run_%d", name, run), func(t *testing.T) {
				options := []tls_client.HttpClientOption{
					skipPeetCertVerify,
					tls_client.WithClientProfile(profile),
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

				var signatureAlgorithms []string

				for _, extension := range tlsApiResponse.TLS.Extensions {
					if strings.HasPrefix(extension.Name, "signature_algorithms (13)") {
						signatureAlgorithms = extension.SignatureAlgorithms
						break
					}
				}

				if len(signatureAlgorithms) == 0 {
					t.Fatalf("%s sent no signature_algorithms extension", name)
				}

				first := signatureAlgorithms[0]

				value, err := strconv.ParseUint(strings.TrimPrefix(first, "0x"), 16, 16)
				if err != nil {
					t.Fatalf("%s first signature algorithm is %q, expected a GREASE value", name, first)
				}

				if !isGreaseValue(value) {
					t.Errorf("%s first signature algorithm is %q, expected a GREASE value", name, first)
				}

				// JA4 ignores GREASE values, so the randomized signature algorithm
				// must not move the JA4 fingerprint.
				if tlsApiResponse.TLS.Ja4 == "" {
					t.Fatalf("%s got no ja4 value from %s", name, peetApiEndpoint)
				}

				t.Logf("%s ja4: %s (first signature algorithm %s)", name, tlsApiResponse.TLS.Ja4, first)

				ja4Values[name] = append(ja4Values[name], tlsApiResponse.TLS.Ja4)
			})
		}
	}

	for name, values := range ja4Values {
		for _, value := range values {
			if value != values[0] {
				t.Errorf("%s ja4 changed between connections: %v", name, values)
				break
			}
		}
	}
}
