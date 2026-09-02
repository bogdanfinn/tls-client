// Package main shows how to build a tls.ClientHelloSpec by hand instead of
// deriving it from a JA3 string. A JA3 string can only name extension IDs, so
// take this route when an extension needs data a JA3 string cannot carry -
// here PSK (pre_shared_key) and ALPS (application_settings).
//
// For the simpler JA3 based route see ./example/custom_profile.
//
// Run: go run ./example/custom_spec
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

// specFactory is called once per ClientHello, so anything that has to stay
// stable for the life of the process belongs outside of it.
func specFactory() (tls.ClientHelloSpec, error) {
	return tls.ClientHelloSpec{
		CipherSuites: []uint16{
			tls.GREASE_PLACEHOLDER,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
		CompressionMethods: []uint8{
			tls.CompressionNone,
		},
		// The order of this slice is the order the extensions go on the wire.
		Extensions: []tls.TLSExtension{
			&tls.UtlsGREASEExtension{},
			&tls.PSKKeyExchangeModesExtension{Modes: []uint8{
				tls.PskModeDHE,
			}},
			&tls.KeyShareExtension{KeyShares: []tls.KeyShare{
				{Group: tls.CurveID(tls.GREASE_PLACEHOLDER), Data: []byte{0}},
				{Group: tls.X25519},
			}},
			// ALPS - a JA3 string cannot express the protocol list.
			&tls.ApplicationSettingsExtension{
				SupportedProtocols: []string{"h2"},
			},
			&tls.SupportedVersionsExtension{Versions: []uint16{
				tls.GREASE_PLACEHOLDER,
				tls.VersionTLS13,
				tls.VersionTLS12,
			}},
			&tls.SNIExtension{},
			&tls.SupportedPointsExtension{SupportedPoints: []byte{
				tls.PointFormatUncompressed,
			}},
			&tls.StatusRequestExtension{},
			&tls.ExtendedMasterSecretExtension{},
			&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
			&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
				tls.CurveID(tls.GREASE_PLACEHOLDER),
				tls.X25519,
				tls.CurveP256,
				tls.CurveP384,
			}},
			&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
			&tls.UtlsCompressCertExtension{Algorithms: []tls.CertCompressionAlgo{
				tls.CertCompressionBrotli,
			}},
			&tls.SCTExtension{},
			&tls.SessionTicketExtension{},
			&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
				tls.ECDSAWithP256AndSHA256,
				tls.PSSWithSHA256,
				tls.PKCS1WithSHA256,
				tls.ECDSAWithP384AndSHA384,
				tls.PSSWithSHA384,
				tls.PKCS1WithSHA384,
				tls.PSSWithSHA512,
				tls.PKCS1WithSHA512,
			}},
			&tls.UtlsGREASEExtension{},
			&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
			// PSK has to be last, and OmitEmptyPsk keeps it off the first
			// handshake where there is no ticket to offer yet.
			&tls.UtlsPreSharedKeyExtension{OmitEmptyPsk: true},
		},
	}, nil
}

func main() {
	settings := map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingEnablePush:           0,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	}
	settingsOrder := []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	}
	pseudoHeaderOrder := []string{":method", ":authority", ":scheme", ":path"}

	customProfile := profiles.NewClientProfile(tls.ClientHelloID{
		Client:      "MyCustomProfileWithPSK",
		Version:     "1",
		SpecFactory: specFactory,
	}, settings, settingsOrder, pseudoHeaderOrder, 15663105, nil, nil, 0, false, nil, nil, 0, nil, false)

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

	req.Header = http.Header{
		"accept":     {"*/*"},
		"user-agent": {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"user-agent",
		},
	}

	// The PSK extension only carries a ticket once the server issued one, so
	// the second request is the one that actually resumes.
	for i := 1; i <= 2; i++ {
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("request %d: status %d\n%s\n", i, resp.StatusCode, body)
	}
}
