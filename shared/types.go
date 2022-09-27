package shared

type TlsApiResponse struct {
	IP          string `json:"ip"`
	HTTPVersion string `json:"http_version"`
	Method      string `json:"method"`
	TLS         struct {
		Ciphers    []string `json:"ciphers"`
		Extensions []struct {
			Name                       string      `json:"name"`
			ServerName                 string      `json:"server_name,omitempty"`
			Data                       string      `json:"data,omitempty"`
			SupportedGroups            []string    `json:"supported_groups,omitempty"`
			EllipticCurvesPointFormats interface{} `json:"elliptic_curves_point_formats,omitempty"`
			Protocols                  []string    `json:"protocols,omitempty"`
			StatusRequest              struct {
				CertificateStatusType   string `json:"certificate_status_type"`
				ResponderIDListLength   int    `json:"responder_id_list_length"`
				RequestExtensionsLength int    `json:"request_extensions_length"`
			} `json:"status_request,omitempty"`
			SignatureAlgorithms []string `json:"signature_algorithms,omitempty"`
			SharedKeys          []struct {
				TLSGrease0X7A7A string `json:"TLS_GREASE (0x7a7a),omitempty"`
				X2551929        string `json:"X25519 (29),omitempty"`
			} `json:"shared_keys,omitempty"`
			PskKeyExchangeMode string   `json:"PSK_Key_Exchange_Mode,omitempty"`
			Versions           []string `json:"versions,omitempty"`
			Algorithms         []string `json:"algorithms,omitempty"`
			PaddingDataLength  int      `json:"padding_data_length,omitempty"`
		} `json:"extensions"`
		TLSVersionRecord     string `json:"tls_version_record"`
		TLSVersionNegotiated string `json:"tls_version_negotiated"`
		Ja3                  string `json:"ja3"`
		Ja3Hash              string `json:"ja3_hash"`
		ClientRandom         string `json:"client_random"`
		SessionID            string `json:"session_id"`
	} `json:"tls"`
	HTTP2 struct {
		AkamaiFingerprint     string `json:"akamai_fingerprint"`
		AkamaiFingerprintHash string `json:"akamai_fingerprint_hash"`
		SentFrames            []struct {
			FrameType string   `json:"frame_type"`
			Length    int      `json:"length"`
			Settings  []string `json:"settings,omitempty"`
			Increment int      `json:"increment,omitempty"`
			StreamID  int      `json:"stream_id,omitempty"`
			Headers   []string `json:"headers,omitempty"`
			Flags     []string `json:"flags,omitempty"`
			Priority  struct {
				Weight    int `json:"weight"`
				DependsOn int `json:"depends_on"`
				Exclusive int `json:"exclusive"`
			} `json:"priority,omitempty"`
		} `json:"sent_frames"`
	} `json:"http2"`
}
