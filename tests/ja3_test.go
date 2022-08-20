package tests

import (
	"crypto/tls"
	"testing"

	utls "github.com/bogdanfinn/utls"
	"github.com/stretchr/testify/assert"

	tls_client "github.com/bogdanfinn/tls-client"
)

func TestJA3_Chrome_104(t *testing.T) {
	input := browserFingerprints[chrome][utls.HelloChrome_104.Str()][ja3String]

	a, err := tls_client.GetClientProfileFromJa3String(input)

	if err != nil {
		t.Fatal(err)
	}

	spec, err := a.GetClientHelloSpec()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, spec.TLSVersMax, uint16(tls.VersionTLS12), "Clients Max TLS Version should be 1.2")
	assert.Equal(t, spec.TLSVersMin, uint16(tls.VersionTLS12), "Clients Min TLS Version should be 1.2")
	assert.Equal(t, len(spec.CipherSuites), 15, "Client should have 15 CipherSuites")
	assert.Equal(t, len(spec.Extensions), 15, "Client should have 15 extensions")
}

func TestJA3_Chrome_103(t *testing.T) {
	input := browserFingerprints[chrome][utls.HelloChrome_103.Str()][ja3String]

	a, err := tls_client.GetClientProfileFromJa3String(input)

	if err != nil {
		t.Fatal(err)
	}

	spec, err := a.GetClientHelloSpec()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, spec.TLSVersMax, uint16(tls.VersionTLS12), "Clients Max TLS Version should be 1.2")
	assert.Equal(t, spec.TLSVersMin, uint16(tls.VersionTLS12), "Clients Min TLS Version should be 1.2")
	assert.Equal(t, len(spec.CipherSuites), 15, "Client should have 15 CipherSuites")
	assert.Equal(t, len(spec.Extensions), 15, "Client should have 15 extensions")
}

func TestJA3_Firefox_102(t *testing.T) {
	input := browserFingerprints[firefox][utls.HelloFirefox_102.Str()][ja3String]

	a, err := tls_client.GetClientProfileFromJa3String(input)

	if err != nil {
		t.Fatal(err)
	}

	spec, err := a.GetClientHelloSpec()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, spec.TLSVersMax, uint16(tls.VersionTLS12), "Clients Max TLS Version should be 1.2")
	assert.Equal(t, spec.TLSVersMin, uint16(tls.VersionTLS12), "Clients Min TLS Version should be 1.2")
	assert.Equal(t, len(spec.CipherSuites), 17, "Client should have 17 CipherSuites")
	assert.Equal(t, len(spec.Extensions), 14, "Client should have 14 extensions")
}

func TestJA3_Opera_89(t *testing.T) {
	input := browserFingerprints[opera][utls.HelloOpera_89.Str()][ja3String]

	a, err := tls_client.GetClientProfileFromJa3String(input)

	if err != nil {
		t.Fatal(err)
	}

	spec, err := a.GetClientHelloSpec()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, spec.TLSVersMax, uint16(tls.VersionTLS12), "Clients Max TLS Version should be 1.2")
	assert.Equal(t, spec.TLSVersMin, uint16(tls.VersionTLS12), "Clients Min TLS Version should be 1.2")
	assert.Equal(t, len(spec.CipherSuites), 15, "Client should have 17 CipherSuites")
	assert.Equal(t, len(spec.Extensions), 15, "Client should have 14 extensions")
}

func TestJA3_Safari_15_5(t *testing.T) {
	input := browserFingerprints[safari][utls.HelloSafari_15_5.Str()][ja3String]

	a, err := tls_client.GetClientProfileFromJa3String(input)

	if err != nil {
		t.Fatal(err)
	}

	spec, err := a.GetClientHelloSpec()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, spec.TLSVersMax, uint16(tls.VersionTLS12), "Clients Max TLS Version should be 1.2")
	assert.Equal(t, spec.TLSVersMin, uint16(tls.VersionTLS12), "Clients Min TLS Version should be 1.2")
	assert.Equal(t, len(spec.CipherSuites), 20, "Client should have 17 CipherSuites")
	assert.Equal(t, len(spec.Extensions), 13, "Client should have 14 extensions")
}
