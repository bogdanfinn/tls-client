package profiles

import (
	"crypto/rand"
	"encoding/binary"

	tls "github.com/bogdanfinn/utls"
)

// randomGREASESignatureScheme returns a random GREASE value for the
// signature_algorithms extension.
//
// Chrome sends a GREASE value as its first signature algorithm and picks a new
// one for every connection. utls replaces GREASE placeholders in the cipher
// suites, the supported groups, the key shares, the supported versions and the
// GREASE extensions, but not in signature_algorithms. The value is therefore
// chosen here. A SpecFactory runs once per connection, so every connection gets
// its own value.
//
// The result has the form 0xWaWa for 0 <= W < 16, which is how BoringSSL builds
// its GREASE values. It panics if the random source fails, which crypto/rand
// does not survive either.
// https://github.com/google/boringssl/blob/master/ssl/handshake_client.cc
//
// This function is a workaround. The proper fix belongs upstream in utls
// (currently github.com/bogdanfinn/utls v1.7.7-barnius) and needs three
// changes:
//
//  1. In u_tls_extensions.go, add a GREASE index for signature algorithms to
//     the const block that starts at ssl_grease_cipher. Put the new name ahead
//     of ssl_grease_ticket_extension so that ssl_grease_last_index, and with it
//     the greaseSeed array on UConn, grows by one.
//  2. In u_parrots.go, add a "case *SignatureAlgorithmsExtension" to the switch
//     over uconn.Extensions inside ApplyPreset, next to the existing
//     SupportedCurvesExtension, KeyShareExtension and SupportedVersionsExtension
//     cases. Replace every entry for which isGREASEUint16 reports true with
//     GetBoringGREASEValue(uconn.greaseSeed, <the new index>), which is what
//     those three cases already do.
//  3. Once utls does the substitution, delete this function and write
//     tls.SignatureScheme(tls.GREASE_PLACEHOLDER) in the profile instead, the
//     same way the cipher suites, the supported groups, the key shares and the
//     supported versions already express GREASE.
//
// The upstream fix is more faithful than this function. utls derives every
// GREASE value in one ClientHello from a single per-connection seed, so the
// values relate to each other the way BoringSSL makes them relate. The value
// returned here is drawn independently of that seed, so it can collide with the
// GREASE value in another position, which real Chrome makes less likely.
func randomGREASESignatureScheme() tls.SignatureScheme {
	var seed [2]byte

	if _, err := rand.Read(seed[:]); err != nil {
		panic(err)
	}

	value := binary.LittleEndian.Uint16(seed[:])
	value = (value & 0xf0) | 0x0a
	value |= value << 8

	return tls.SignatureScheme(value)
}
