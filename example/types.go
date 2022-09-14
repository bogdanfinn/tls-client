package main

type TlsApiResponse struct {
	Donate      string `json:"donate"`
	IP          string `json:"ip"`
	HTTPVersion string `json:"http_version"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	TLS         struct {
		Ciphers        []string `json:"ciphers"`
		Curves         []string `json:"curves"`
		Extensions     []string `json:"extensions"`
		Points         []string `json:"points"`
		Version        string   `json:"version"`
		Protocols      []string `json:"protocols"`
		Versions       []string `json:"versions"`
		Ja3            string   `json:"ja3"`
		Ja3Hash        string   `json:"ja3_hash"`
		Ja3Padding     string   `json:"ja3_padding"`
		Ja3HashPadding string   `json:"ja3_hash_padding"`
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
