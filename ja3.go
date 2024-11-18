package tls_client

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"

	tls "github.com/bogdanfinn/utls"
)

type CandidateCipherSuites struct {
	KdfId  string
	AeadId string
}

func GetSpecFactoryFromJa3String(ja3String string, supportedSignatureAlgorithms, supportedDelegatedCredentialsAlgorithms, supportedVersions, keyShareCurves, supportedProtocolsALPN, supportedProtocolsALPS []string, echCandidateCipherSuites []CandidateCipherSuites, candidatePayloads []uint16, certCompressionAlgo string) (func() (tls.ClientHelloSpec, error), error) {
	return func() (tls.ClientHelloSpec, error) {
		var mappedSignatureAlgorithms []tls.SignatureScheme

		for _, supportedSignatureAlgorithm := range supportedSignatureAlgorithms {
			signatureAlgorithm, ok := signatureAlgorithms[supportedSignatureAlgorithm]
			if ok {
				mappedSignatureAlgorithms = append(mappedSignatureAlgorithms, signatureAlgorithm)
			} else {
				supportedSignatureAlgorithmAsUint, err := strconv.ParseUint(supportedSignatureAlgorithm, 16, 16)

				if err != nil {
					return tls.ClientHelloSpec{}, fmt.Errorf("%s is not a valid supportedSignatureAlgorithm", supportedSignatureAlgorithm)
				}

				mappedSignatureAlgorithms = append(mappedSignatureAlgorithms, tls.SignatureScheme(uint16(supportedSignatureAlgorithmAsUint)))
			}
		}

		var mappedDelegatedCredentialsAlgorithms []tls.SignatureScheme

		for _, supportedDelegatedCredentialsAlgorithm := range supportedDelegatedCredentialsAlgorithms {
			delegatedCredentialsAlgorithm, ok := delegatedCredentialsAlgorithms[supportedDelegatedCredentialsAlgorithm]
			if ok {
				mappedDelegatedCredentialsAlgorithms = append(mappedDelegatedCredentialsAlgorithms, delegatedCredentialsAlgorithm)
			} else {
				supportedDelegatedCredentialsAlgorithmAsUint, err := strconv.ParseUint(supportedDelegatedCredentialsAlgorithm, 16, 16)

				if err != nil {
					return tls.ClientHelloSpec{}, fmt.Errorf("%s is not a valid supportedDelegatedCredentialsAlgorithm", supportedDelegatedCredentialsAlgorithm)
				}

				mappedDelegatedCredentialsAlgorithms = append(mappedDelegatedCredentialsAlgorithms, tls.SignatureScheme(uint16(supportedDelegatedCredentialsAlgorithmAsUint)))
			}
		}

		var mappedHpkeSymmetricCipherSuites []tls.HPKESymmetricCipherSuite

		for _, echCandidateCipherSuites := range echCandidateCipherSuites {
			kdfId, ok1 := kdfIds[echCandidateCipherSuites.KdfId]

			aeadId, ok2 := aeadIds[echCandidateCipherSuites.AeadId]
			if ok1 && ok2 {
				mappedHpkeSymmetricCipherSuites = append(mappedHpkeSymmetricCipherSuites, tls.HPKESymmetricCipherSuite{
					KdfId:  kdfId,
					AeadId: aeadId,
				})
			} else {
				kdfId, err := strconv.ParseUint(echCandidateCipherSuites.KdfId, 16, 16)
				if err != nil {
					return tls.ClientHelloSpec{}, fmt.Errorf("%s is not a valid KdfId", echCandidateCipherSuites.KdfId)
				}

				aeadId, err := strconv.ParseUint(echCandidateCipherSuites.AeadId, 16, 16)
				if err != nil {
					return tls.ClientHelloSpec{}, fmt.Errorf("%s is not a valid aeadId", echCandidateCipherSuites.AeadId)
				}

				mappedHpkeSymmetricCipherSuites = append(mappedHpkeSymmetricCipherSuites, tls.HPKESymmetricCipherSuite{
					KdfId:  uint16(kdfId),
					AeadId: uint16(aeadId),
				})
			}
		}

		var mappedTlsVersions []uint16

		for _, version := range supportedVersions {
			mappedVersion, ok := tlsVersions[version]
			if ok {
				mappedTlsVersions = append(mappedTlsVersions, mappedVersion)
			}
		}

		var mappedKeyShares []tls.KeyShare

		for _, keyShareCurve := range keyShareCurves {
			resolvedKeyShare, ok := curves[keyShareCurve]

			if !ok {
				continue
			}

			mappedKeyShare := tls.KeyShare{Group: resolvedKeyShare}

			if keyShareCurve == "GREASE" {
				mappedKeyShare.Data = []byte{0}
			}

			mappedKeyShares = append(mappedKeyShares, mappedKeyShare)
		}

		compressionAlgo, ok := certCompression[certCompressionAlgo]

		if !ok {
			return stringToSpec(ja3String, mappedSignatureAlgorithms, mappedDelegatedCredentialsAlgorithms, mappedTlsVersions, mappedKeyShares, mappedHpkeSymmetricCipherSuites, candidatePayloads, supportedProtocolsALPN, supportedProtocolsALPS, nil)
		}

		return stringToSpec(ja3String, mappedSignatureAlgorithms, mappedDelegatedCredentialsAlgorithms, mappedTlsVersions, mappedKeyShares, mappedHpkeSymmetricCipherSuites, candidatePayloads, supportedProtocolsALPN, supportedProtocolsALPS, &compressionAlgo)
	}, nil
}

func stringToSpec(ja3 string, signatureAlgorithms []tls.SignatureScheme, delegatedCredentialsAlgorithms []tls.SignatureScheme, tlsVersions []uint16, keyShares []tls.KeyShare, hpkeSymmetricCipherSuites []tls.HPKESymmetricCipherSuite, candidatePayloads []uint16, supportedProtocolsALPN, supportedProtocolsALPS []string, certCompression *tls.CertCompressionAlgo) (tls.ClientHelloSpec, error) {
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

	if certCompression == nil && strings.Contains(ja3StringParts[2], fmt.Sprintf("%d", tls.ExtensionCompressCertificate)) {
		fmt.Println("attention our ja3 defines ExtensionCompressCertificate but you did not specify certCompression")
	}

	if certCompression != nil {
		extMap[tls.ExtensionCompressCertificate] = &tls.UtlsCompressCertExtension{Algorithms: []tls.CertCompressionAlgo{*certCompression}}
	}

	extMap[tls.ExtensionKeyShare] = &tls.KeyShareExtension{KeyShares: keyShares}
	extMap[tls.ExtensionSupportedPoints] = &tls.SupportedPointsExtension{SupportedPoints: targetPointFormats}
	extMap[tls.ExtensionECH] = &tls.GREASEEncryptedClientHelloExtension{
		CandidateCipherSuites: hpkeSymmetricCipherSuites,
		CandidatePayloadLens:  candidatePayloads,
	}
	extMap[tls.ExtensionSupportedVersions] = &tls.SupportedVersionsExtension{Versions: tlsVersions}
	extMap[tls.ExtensionSignatureAlgorithms] = &tls.SignatureAlgorithmsExtension{
		SupportedSignatureAlgorithms: signatureAlgorithms,
	}

	extMap[tls.ExtensionDelegatedCredentials] = &tls.DelegatedCredentialsExtension{
		SupportedSignatureAlgorithms: delegatedCredentialsAlgorithms,
	}

	extMap[tls.ExtensionALPN] = &tls.ALPNExtension{
		AlpnProtocols: supportedProtocolsALPN,
	}

	extMap[tls.ExtensionALPS] = &tls.ApplicationSettingsExtension{
		SupportedProtocols: supportedProtocolsALPS,
	}

	var exts []tls.TLSExtension
	for _, e := range extensions {
		eId, err := strconv.ParseUint(e, 10, 16)

		if err != nil {
			return tls.ClientHelloSpec{}, err
		}

		if uint16(eId) == tls.GREASE_PLACEHOLDER {
			// if we use multiple grease extensions with need to generate always a new value. therefore we are creating a new instance here
			exts = append(exts, &tls.UtlsGREASEExtension{})
			continue
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
		CompressionMethods: []byte{tls.CompressionNone},
		Extensions:         exts,
		GetSessionID:       sha256.Sum256,
	}, nil
}

func getExtensionBaseMap() map[uint16]tls.TLSExtension {
	return map[uint16]tls.TLSExtension{
		// This extension needs to be instantiated every time and not be reused if it occurs multiple times in the same ja3
		//tls.GREASE_PLACEHOLDER:     &tls.UtlsGREASEExtension{},

		tls.ExtensionServerName:    &tls.SNIExtension{},
		tls.ExtensionStatusRequest: &tls.StatusRequestExtension{},

		// These are applied later
		// tls.ExtensionSupportedCurves: &tls.SupportedCurvesExtension{...}
		// tls.ExtensionSupportedPoints: &tls.SupportedPointsExtension{...}
		// tls.ExtensionSignatureAlgorithms: &tls.SignatureAlgorithmsExtension{...}
		// tls.ExtensionCompressCertificate:  &tls.UtlsCompressCertExtension{...},
		// tls.ExtensionSupportedVersions: &tls.SupportedVersionsExtension{...}
		// tls.ExtensionKeyShare:     &tls.KeyShareExtension{...},
		// tls.ExtensionDelegatedCredentials: &tls.DelegatedCredentialsExtension{},
		// tls.ExtensionALPN: &tls.ALPNExtension{},
		// tls.ExtensionALPS:         &tls.ApplicationSettingsExtension{},

		tls.ExtensionSCT:                  &tls.SCTExtension{},
		tls.ExtensionPadding:              &tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
		tls.ExtensionExtendedMasterSecret: &tls.ExtendedMasterSecretExtension{},
		tls.ExtensionRecordSizeLimit:      &tls.FakeRecordSizeLimitExtension{},
		tls.ExtensionSessionTicket:        &tls.SessionTicketExtension{},
		tls.ExtensionPreSharedKey:         &tls.UtlsPreSharedKeyExtension{},
		tls.ExtensionEarlyData:            &tls.GenericExtension{Id: tls.ExtensionEarlyData},
		tls.ExtensionCookie:               &tls.CookieExtension{},
		tls.ExtensionPSKModes: &tls.PSKKeyExchangeModesExtension{
			Modes: []uint8{
				tls.PskModeDHE,
			}},
		tls.ExtensionNextProtoNeg: &tls.NPNExtension{},
		tls.ExtensionRenegotiationInfo: &tls.RenegotiationInfoExtension{
			Renegotiation: tls.RenegotiateOnceAsClient,
		},
	}
}
