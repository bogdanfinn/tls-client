package tests

import (
	"testing"

	utls "github.com/bogdanfinn/utls"
	"github.com/stretchr/testify/assert"

	tls_client "github.com/bogdanfinn/tls-client"
)

func TestJA3(t *testing.T) {
	t.Log("testing ja3 chrome 105")
	ja3_chrome_105(t)
	t.Log("testing ja3 chrome 107")
	ja3_chrome_107(t)
	t.Log("testing ja3 firefox")
	ja3_firefox_105(t)
	t.Log("testing ja3 opera")
	ja3_opera_91(t)
}

func ja3_chrome_105(t *testing.T) {
	input := clientFingerprints[chrome][utls.HelloChrome_105.Str()][ja3String]

	ssa := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	dca := []string{"PKCS1WithSHA256", "PKCS1WithSHA384", "PKCS1WithSHA512"}
	sv := []string{"1.3", "1.2"}
	sc := []string{"GREASE", "X25519"}

	specFunc, err := tls_client.GetSpecFactoryFromJa3String(input, ssa, dca, sv, sc, "zlib")

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

	specFunc, err := tls_client.GetSpecFactoryFromJa3String(input, ssa, dca, sv, sc, "zlib")

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

	specFunc, err := tls_client.GetSpecFactoryFromJa3String(input, ssa, dca, sv, sc, "zlib")

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

	specFunc, err := tls_client.GetSpecFactoryFromJa3String(input, ssa, dca, sv, sc, "zlib")

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
