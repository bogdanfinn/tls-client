package tests

import (
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"
	utls "github.com/bogdanfinn/utls"
	"github.com/stretchr/testify/assert"

	tls_client "github.com/bogdanfinn/tls-client"
)

func TestJA3(t *testing.T) {
	t.Log("testing ja3 chrome 120")
	ja3_chrome_120(t)
	t.Log("testing ja3 chrome 116 with psk")
	ja3_chrome_112_with_psk(t)
	t.Log("testing ja3 chrome 105")
	ja3_chrome_105(t)
	t.Log("testing ja3 chrome 107")
	ja3_chrome_107(t)
	t.Log("testing ja3 firefox")
	ja3_firefox_105(t)
	t.Log("testing ja3 opera")
	ja3_opera_91(t)
}

func ja3_chrome_120(t *testing.T) {
	input := clientFingerprints[chrome][profiles.Chrome_120.GetClientHelloStr()][ja3String]

	ssa := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	dca := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	sv := []string{"1.3", "1.2"}
	sc := []string{"GREASE", "X25519"}
	alpnProtocols := []string{"h2", "http/1.1"}
	alpsProtocols := []string{"h2"}

	ccs := []tls_client.CandidateCipherSuites{
		{
			KdfId:  "HKDF_SHA256",
			AeadId: "AEAD_AES_128_GCM",
		},
		{
			KdfId:  "HKDF_SHA256",
			AeadId: "AEAD_CHACHA20_POLY1305",
		},
	}
	cp := []uint16{128, 160, 192, 224}

	specFunc, err := tls_client.GetSpecFactoryFromJa3String(input, ssa, dca, sv, sc, alpnProtocols, alpsProtocols, ccs, cp, "zlib")

	if err != nil {
		t.Fatal(err)
	}

	spec, err := specFunc()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(spec.CipherSuites), 15, "Client should have 15 CipherSuites")
	assert.Equal(t, len(spec.Extensions), 16, "Client should have 16 extensions")
}

func ja3_chrome_112_with_psk(t *testing.T) {
	input := clientFingerprints[chrome][utls.HelloChrome_112_PSK.Str()][ja3String]

	ssa := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	dca := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	sv := []string{"1.3", "1.2"}
	sc := []string{"GREASE", "X25519"}
	alpnProtocols := []string{"h2", "http/1.1"}
	alpsProtocols := []string{"h2"}

	specFunc, err := tls_client.GetSpecFactoryFromJa3String(input, ssa, dca, sv, sc, alpnProtocols, alpsProtocols, nil, nil, "brotli")

	if err != nil {
		t.Fatal(err)
	}

	spec, err := specFunc()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 15, len(spec.CipherSuites), "Client should have 15 CipherSuites")
	assert.Equal(t, 17, len(spec.Extensions), "Client should have 17 extensions")
}

func ja3_chrome_105(t *testing.T) {
	input := clientFingerprints[chrome][utls.HelloChrome_105.Str()][ja3String]

	ssa := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	dca := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	sv := []string{"1.3", "1.2"}
	sc := []string{"GREASE", "X25519"}
	alpnProtocols := []string{"h2", "http/1.1"}
	alpsProtocols := []string{"h2"}

	specFunc, err := tls_client.GetSpecFactoryFromJa3String(input, ssa, dca, sv, sc, alpnProtocols, alpsProtocols, nil, nil, "zlib")

	if err != nil {
		t.Fatal(err)
	}

	spec, err := specFunc()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(spec.CipherSuites), 15, "Client should have 15 CipherSuites")
	assert.Equal(t, len(spec.Extensions), 16, "Client should have 16 extensions")
}

func ja3_chrome_107(t *testing.T) {
	input := clientFingerprints[chrome][utls.HelloChrome_107.Str()][ja3String]

	ssa := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	dca := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	sv := []string{"1.3", "1.2"}
	sc := []string{"GREASE", "X25519"}
	alpnProtocols := []string{"h2", "http/1.1"}
	alpsProtocols := []string{"h2"}

	specFunc, err := tls_client.GetSpecFactoryFromJa3String(input, ssa, dca, sv, sc, alpnProtocols, alpsProtocols, nil, nil, "zlib")

	if err != nil {
		t.Fatal(err)
	}

	spec, err := specFunc()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(spec.CipherSuites), 15, "Client should have 15 CipherSuites")
	assert.Equal(t, len(spec.Extensions), 16, "Client should have 16 extensions")
}

func ja3_firefox_105(t *testing.T) {
	input := clientFingerprints[firefox][utls.HelloFirefox_105.Str()][ja3String]

	ssa := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	dca := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	sv := []string{"1.3", "1.2"}
	sc := []string{"GREASE", "X25519"}
	alpnProtocols := []string{"h2", "http/1.1"}
	alpsProtocols := []string{"h2"}

	specFunc, err := tls_client.GetSpecFactoryFromJa3String(input, ssa, dca, sv, sc, alpnProtocols, alpsProtocols, nil, nil, "zlib")

	if err != nil {
		t.Fatal(err)
	}

	spec, err := specFunc()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(spec.CipherSuites), 17, "Client should have 17 CipherSuites")
	assert.Equal(t, len(spec.Extensions), 15, "Client should have 15 extensions")
}

func ja3_opera_91(t *testing.T) {
	input := clientFingerprints[opera][utls.HelloOpera_91.Str()][ja3String]

	ssa := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	dca := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	sv := []string{"1.3", "1.2"}
	sc := []string{"GREASE", "X25519"}
	alpnProtocols := []string{"h2", "http/1.1"}
	alpsProtocols := []string{"h2"}

	specFunc, err := tls_client.GetSpecFactoryFromJa3String(input, ssa, dca, sv, sc, alpnProtocols, alpsProtocols, nil, nil, "zlib")

	if err != nil {
		t.Fatal(err)
	}

	spec, err := specFunc()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(spec.CipherSuites), 15, "Client should have 15 CipherSuites")
	assert.Equal(t, len(spec.Extensions), 16, "Client should have 16 extensions")
}
