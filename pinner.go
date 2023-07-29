package tls_client

import (
	"errors"
	"fmt"
	"strings"

	http "github.com/bogdanfinn/fhttp"
	tls "github.com/bogdanfinn/utls"
	"github.com/tam7t/hpkp"
)

var DefaultBadPinHandler = func(req *http.Request) {
	fmt.Println("this is the default bad pin handler")
}

var ErrBadPinDetected = errors.New("bad ssl pin detected")

var certificatePinStorage = hpkp.NewMemStorage()

type certificatePinner struct {
	certificatePins map[string][]string
}

type CertificatePinner interface {
	Pin(conn *tls.UConn, host string) error
}

func NewCertificatePinner(certificatePins map[string][]string) (CertificatePinner, error) {
	pinner := &certificatePinner{
		certificatePins: certificatePins,
	}

	err := pinner.init()
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate certificate pinner: %w", err)
	}

	return pinner, nil
}

func (cp *certificatePinner) init() error {
	if len(cp.certificatePins) == 0 {
		return nil
	}

	for host, pinsByHost := range cp.certificatePins {
		includeSubdomains := false
		if strings.Contains(host, "*.") {
			includeSubdomains = true
			host = strings.ReplaceAll(host, "*.", "")
		}

		pinnedHost := certificatePinStorage.Lookup(host)

		if pinnedHost != nil {
			// another pinner instance already initialized the host. we do not need to pin again
			continue
		}

		certificatePinStorage.Add(host, &hpkp.Header{
			Permanent:         true,
			Sha256Pins:        pinsByHost,
			IncludeSubDomains: includeSubdomains,
		})
	}

	return nil
}

func (cp *certificatePinner) Pin(conn *tls.UConn, host string) error {
	validPin := false

	if len(cp.certificatePins) == 0 {
		return nil
	}

	pinnedHost := certificatePinStorage.Lookup(host)

	if pinnedHost == nil {
		// host is not pinned, we treat it as valid
		return nil
	}

	for _, peerCert := range conn.ConnectionState().PeerCertificates {
		peerPin := hpkp.Fingerprint(peerCert)

		if pinnedHost.Matches(peerPin) {
			validPin = true
			break
		}
	}

	if !validPin {
		return ErrBadPinDetected
	}

	return nil
}
