package profiles

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"os"
	"testing"

	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
	"github.com/stretchr/testify/require"
)

// Compare a serialized ClientHello and the HTTP/2 configuration to a real
// browser capture. The duplicate PSS SHA384 signature algorithm is intentional.
func TestSafari272Capture(t *testing.T) {
	var capture struct {
		TLS struct {
			CipherSuites []struct{ ID uint16 } `json:"cipher_suites"`
			Extensions   []struct {
				ID   uint16
				Data struct {
					NamedGroupList               []struct{ ID uint16 } `json:"named_group_list"`
					ECPointFormatList            []struct{ ID uint8 }  `json:"ec_point_format_list"`
					SupportedSignatureAlgorithms []struct{ ID uint16 } `json:"supported_signature_algorithms"`
					Versions                     []struct{ ID uint16 }
					Algorithms                   []struct{ ID uint16 }
					ProtocolNameList             []string             `json:"protocol_name_list"`
					KeModes                      []struct{ ID uint8 } `json:"ke_modes"`
					ClientShares                 []struct {
						Group             struct{ ID uint16 }
						KeyExchangeLength int `json:"key_exchange_length"`
					} `json:"client_shares"`
				}
			}
		} `json:"tls"`
		HTTP2 []struct {
			ID       int
			Settings []struct {
				ID    http2.SettingID
				Value uint32
			}
			WindowSizeIncrement uint32 `json:"window_size_increment"`
		} `json:"http2"`
	}
	data, err := os.ReadFile("testdata/safari_27_2.json")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &capture))
	profile, ok := MappedTLSClients["safari_27_2"]
	require.True(t, ok)
	require.Equal(t, "Safari-27.2", profile.GetClientHelloStr())
	spec, err := profile.GetClientHelloSpec()
	require.NoError(t, err)
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	conn := tls.UClient(client, &tls.Config{ServerName: "tls.browserleaks.com"}, tls.HelloCustom, false, false, false)
	require.NoError(t, conn.ApplyPreset(&spec))
	require.NoError(t, conn.BuildHandshakeState())
	// Inspect the actual marshalled ClientHello, including generated key shares.
	hello := conn.HandshakeState.Hello.Raw
	require.GreaterOrEqual(t, len(hello), 39)
	require.Equal(t, uint16(771), binary.BigEndian.Uint16(hello[4:6]))
	offset := 39 + int(hello[38]) // handshake header, version, random, session ID
	cipherLen := int(binary.BigEndian.Uint16(hello[offset:]))
	offset += 2
	require.Equal(t, len(capture.TLS.CipherSuites)*2, cipherLen)
	for i, cipher := range capture.TLS.CipherSuites {
		require.Equal(t, safariNormalizeGrease(cipher.ID), safariNormalizeGrease(binary.BigEndian.Uint16(hello[offset+2*i:])))
	}
	offset += cipherLen
	compressionLen := int(hello[offset])
	require.Equal(t, []byte{0}, hello[offset+1:offset+1+compressionLen])
	offset += 1 + compressionLen
	extensionLen := int(binary.BigEndian.Uint16(hello[offset:]))
	offset += 2
	require.Equal(t, len(hello), offset+extensionLen)
	for _, ext := range capture.TLS.Extensions {
		require.GreaterOrEqual(t, len(hello)-offset, 4)
		id := binary.BigEndian.Uint16(hello[offset:])
		size := int(binary.BigEndian.Uint16(hello[offset+2:]))
		offset += 4
		require.GreaterOrEqual(t, len(hello)-offset, size)
		body := hello[offset : offset+size]
		offset += size
		require.Equal(t, safariNormalizeGrease(ext.ID), safariNormalizeGrease(id))
		switch id {
		case 10, 13, 43, 27:
			var expected []uint16
			skip := 2
			switch id {
			case 10:
				for _, v := range ext.Data.NamedGroupList {
					expected = append(expected, safariNormalizeGrease(v.ID))
				}
			case 13:
				for _, v := range ext.Data.SupportedSignatureAlgorithms {
					expected = append(expected, v.ID)
				}
			case 43:
				skip = 1
				for _, v := range ext.Data.Versions {
					expected = append(expected, safariNormalizeGrease(v.ID))
				}
			case 27:
				skip = 1
				for _, v := range ext.Data.Algorithms {
					expected = append(expected, v.ID)
				}
			}
			var actual []uint16
			for i := skip; i < len(body); i += 2 {
				actual = append(actual, safariNormalizeGrease(binary.BigEndian.Uint16(body[i:])))
			}
			require.Equal(t, expected, actual, "extension %d", id)
		case 11, 45:
			expected := []byte{}
			if id == 11 {
				for _, v := range ext.Data.ECPointFormatList {
					expected = append(expected, v.ID)
				}
			} else {
				for _, v := range ext.Data.KeModes {
					expected = append(expected, v.ID)
				}
			}
			require.Equal(t, expected, body[1:])
		case 16:
			var actual []string
			for i := 2; i < len(body); {
				n := int(body[i])
				actual = append(actual, string(body[i+1:i+1+n]))
				i += 1 + n
			}
			require.Equal(t, ext.Data.ProtocolNameList, actual)
		case 51:
			i := 2
			for _, share := range ext.Data.ClientShares {
				require.GreaterOrEqual(t, len(body)-i, 4)
				require.Equal(t, safariNormalizeGrease(share.Group.ID), safariNormalizeGrease(binary.BigEndian.Uint16(body[i:])))
				n := int(binary.BigEndian.Uint16(body[i+2:]))
				require.Equal(t, share.KeyExchangeLength, n)
				i += 4 + n
			}
			require.Equal(t, len(body), i)
		}
	}
	require.Equal(t, len(hello), offset)
	for _, frame := range capture.HTTP2 {
		switch frame.ID {
		case 4:
			var order []http2.SettingID
			settings := map[http2.SettingID]uint32{}
			for _, s := range frame.Settings {
				order = append(order, s.ID)
				settings[s.ID] = s.Value
			}
			require.Equal(t, order, profile.GetSettingsOrder())
			require.Equal(t, settings, profile.GetSettings())
		case 8:
			require.Equal(t, frame.WindowSizeIncrement, profile.GetConnectionFlow())
		}
	}
	require.Equal(t, []string{":method", ":scheme", ":authority", ":path"}, profile.GetPseudoHeaderOrder())
	require.Empty(t, profile.GetPriorities())
	require.Nil(t, profile.GetHeaderPriority())
}

func safariNormalizeGrease(value uint16) uint16 {
	if value&0x0f0f == 0x0a0a && value>>8 == value&0xff {
		return tls.GREASE_PLACEHOLDER
	}
	return value
}
