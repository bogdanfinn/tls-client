package tls_client

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"

	tls "github.com/bogdanfinn/utls"
)

func GetSpecFactorFromJa3String(ja3String string) (func() (tls.ClientHelloSpec, error), error) {
	return func() (tls.ClientHelloSpec, error) {
		spec, err := stringToSpec(ja3String)

		return spec, err
	}, nil
}

func stringToSpec(ja3 string) (tls.ClientHelloSpec, error) {
	extMap := getExtensionBaseMap()
	ja3StringParts := strings.Split(ja3, ",")

	ciphers := strings.Split(ja3StringParts[1], "-")
	extensions := strings.Split(ja3StringParts[2], "-")
	curves := strings.Split(ja3StringParts[3], "-")

	if len(curves) == 1 && curves[0] == "" {
		curves = []string{}
	}

	pointFormats := strings.Split(ja3StringParts[4], "-")
	if len(pointFormats) == 1 && pointFormats[0] == "" {
		pointFormats = []string{}
	}

	var targetCurves []tls.CurveID
	for _, c := range curves {
		cid, err := strconv.ParseUint(c, 10, 16)
		if err != nil {
			return tls.ClientHelloSpec{}, err
		}
		targetCurves = append(targetCurves, tls.CurveID(cid))
	}

	extMap[tls.ExtensionSupportedCurves] = &tls.SupportedCurvesExtension{Curves: targetCurves}

	// parse point formats
	var targetPointFormats []byte
	for _, p := range pointFormats {
		pid, err := strconv.ParseUint(p, 10, 8)
		if err != nil {
			return tls.ClientHelloSpec{}, err
		}
		targetPointFormats = append(targetPointFormats, byte(pid))
	}

	extMap[tls.ExtensionSupportedPoints] = &tls.SupportedPointsExtension{SupportedPoints: targetPointFormats}

	var exts []tls.TLSExtension
	for _, e := range extensions {
		eId, err := strconv.ParseUint(e, 10, 16)

		if err != nil {
			return tls.ClientHelloSpec{}, err
		}

		te, ok := extMap[uint16(eId)]
		if !ok {
			return tls.ClientHelloSpec{}, fmt.Errorf("unknown extension with id %s provided", e)
		}
		exts = append(exts, te)
	}

	var suites []uint16
	for _, c := range ciphers {
		cid, err := strconv.ParseUint(c, 10, 16)
		if err != nil {
			return tls.ClientHelloSpec{}, err
		}
		suites = append(suites, uint16(cid))
	}

	return tls.ClientHelloSpec{
		CipherSuites:       suites,
		CompressionMethods: []byte{0},
		Extensions:         exts,
		GetSessionID:       sha256.Sum256,
	}, nil
}

func getExtensionBaseMap() map[uint16]tls.TLSExtension {
	return map[uint16]tls.TLSExtension{
		tls.ExtensionServerName:    &tls.SNIExtension{},
		tls.ExtensionStatusRequest: &tls.StatusRequestExtension{},

		// These are applied later
		// tls.ExtensionSupportedCurves: &tls.SupportedCurvesExtension{...}
		// tls.ExtensionSupportedPoints: &tls.SupportedPointsExtension{...}

		tls.ExtensionSignatureAlgorithms: &tls.SignatureAlgorithmsExtension{
			SupportedSignatureAlgorithms: []tls.SignatureScheme{
				tls.ECDSAWithP256AndSHA256,
				tls.PSSWithSHA256,
				tls.PKCS1WithSHA256,
				tls.ECDSAWithP384AndSHA384,
				tls.PSSWithSHA384,
				tls.PKCS1WithSHA384,
				tls.PSSWithSHA512,
				tls.PKCS1WithSHA512,
				// tls.PKCS1WithSHA1,
			},
		},
		tls.ExtensionALPN: &tls.ALPNExtension{
			AlpnProtocols: []string{"h2", "http/1.1"},
		},
		tls.ExtensionSCT:                  &tls.SCTExtension{},
		tls.ExtensionPadding:              &tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
		tls.ExtensionExtendedMasterSecret: &tls.UtlsExtendedMasterSecretExtension{},
		tls.ExtensionCompressCertificate:  &tls.UtlsCompressCertExtension{},
		tls.ExtensionRecordSizeLimit:      &tls.FakeRecordSizeLimitExtension{},
		tls.ExtensionDelegatedCredentials: &tls.DelegatedCredentialsExtension{},
		tls.ExtensionSessionTicket:        &tls.SessionTicketExtension{},
		tls.ExtensionPreSharedKey:         &tls.PreSharedKeyExtension{},
		tls.ExtensionEarlyData:            &tls.GenericExtension{Id: tls.ExtensionEarlyData},
		tls.ExtensionSupportedVersions: &tls.SupportedVersionsExtension{Versions: []uint16{
			// tls.GREASE_PLACEHOLDER,
			tls.VersionTLS13,
			tls.VersionTLS12,
			tls.VersionTLS11,
			tls.VersionTLS10}},
		tls.ExtensionCookie: &tls.CookieExtension{},
		tls.ExtensionPSKModes: &tls.PSKKeyExchangeModesExtension{
			Modes: []uint8{
				tls.PskModeDHE,
			}},
		tls.ExtensionKeyShare:     &tls.KeyShareExtension{KeyShares: []tls.KeyShare{{Group: tls.X25519}}},
		tls.ExtensionNextProtoNeg: &tls.NPNExtension{},
		tls.ExtensionALPS:         &tls.ALPSExtension{},
		tls.ExtensionRenegotiationInfo: &tls.RenegotiationInfoExtension{
			Renegotiation: tls.RenegotiateOnceAsClient,
		},
	}
}
