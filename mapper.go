package tls_client

import (
	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
	"github.com/bogdanfinn/utls/dicttls"
)

var H2SettingsMap = map[string]http2.SettingID{
	"HEADER_TABLE_SIZE":      http2.SettingHeaderTableSize,
	"ENABLE_PUSH":            http2.SettingEnablePush,
	"MAX_CONCURRENT_STREAMS": http2.SettingMaxConcurrentStreams,
	"INITIAL_WINDOW_SIZE":    http2.SettingInitialWindowSize,
	"MAX_FRAME_SIZE":         http2.SettingMaxFrameSize,
	"MAX_HEADER_LIST_SIZE":   http2.SettingMaxHeaderListSize,
	"UNKNOWN_SETTING_7":      0x7,
	"UNKNOWN_SETTING_8":      0x8,
	"UNKNOWN_SETTING_9":      0x9,
}

var tlsVersions = map[string]uint16{
	"GREASE": tls.GREASE_PLACEHOLDER,
	"1.3":    tls.VersionTLS13,
	"1.2":    tls.VersionTLS12,
	"1.1":    tls.VersionTLS11,
	"1.0":    tls.VersionTLS10,
}

var signatureAlgorithms = map[string]tls.SignatureScheme{
	"PKCS1WithSHA256":        tls.PKCS1WithSHA256,
	"PKCS1WithSHA384":        tls.PKCS1WithSHA384,
	"PKCS1WithSHA512":        tls.PKCS1WithSHA512,
	"PSSWithSHA256":          tls.PSSWithSHA256,
	"PSSWithSHA384":          tls.PSSWithSHA384,
	"PSSWithSHA512":          tls.PSSWithSHA512,
	"ECDSAWithP256AndSHA256": tls.ECDSAWithP256AndSHA256,
	"ECDSAWithP384AndSHA384": tls.ECDSAWithP384AndSHA384,
	"ECDSAWithP521AndSHA512": tls.ECDSAWithP521AndSHA512,
	"PKCS1WithSHA1":          tls.PKCS1WithSHA1,
	"ECDSAWithSHA1":          tls.ECDSAWithSHA1,
	"Ed25519":                tls.Ed25519,
	"SHA224_RSA":             tls.SHA224_RSA,
	"SHA224_ECDSA":           tls.SHA224_ECDSA,
}

var delegatedCredentialsAlgorithms = map[string]tls.SignatureScheme{
	"PKCS1WithSHA256":        tls.PKCS1WithSHA256,
	"PKCS1WithSHA384":        tls.PKCS1WithSHA384,
	"PKCS1WithSHA512":        tls.PKCS1WithSHA512,
	"PSSWithSHA256":          tls.PSSWithSHA256,
	"PSSWithSHA384":          tls.PSSWithSHA384,
	"PSSWithSHA512":          tls.PSSWithSHA512,
	"ECDSAWithP256AndSHA256": tls.ECDSAWithP256AndSHA256,
	"ECDSAWithP384AndSHA384": tls.ECDSAWithP384AndSHA384,
	"ECDSAWithP521AndSHA512": tls.ECDSAWithP521AndSHA512,
	"PKCS1WithSHA1":          tls.PKCS1WithSHA1,
	"ECDSAWithSHA1":          tls.ECDSAWithSHA1,
	"Ed25519":                tls.Ed25519,
}

var kdfIds = map[string]uint16{
	"HKDF_SHA256": dicttls.HKDF_SHA256,
	"HKDF_SHA384": dicttls.HKDF_SHA384,
	"HKDF_SHA512": dicttls.HKDF_SHA512,
}

var aeadIds = map[string]uint16{
	"AEAD_AES_128_GCM":       dicttls.AEAD_AES_128_GCM,
	"AEAD_AES_256_GCM":       dicttls.AEAD_AES_256_GCM,
	"AEAD_CHACHA20_POLY1305": dicttls.AEAD_CHACHA20_POLY1305,
}

var curves = map[string]tls.CurveID{
	"GREASE":          tls.CurveID(tls.GREASE_PLACEHOLDER),
	"P256":            tls.CurveP256,
	"P384":            tls.CurveP384,
	"P521":            tls.CurveP521,
	"X25519":          tls.X25519,
	"P256Kyber768":    tls.P256Kyber768Draft00,
	"X25519Kyber512D": tls.X25519Kyber512Draft00,
	"X25519Kyber768":  tls.X25519Kyber768Draft00,
}

var certCompression = map[string]tls.CertCompressionAlgo{
	"zlib":   tls.CertCompressionZlib,
	"brotli": tls.CertCompressionBrotli,
	"zstd":   tls.CertCompressionZstd,
}
