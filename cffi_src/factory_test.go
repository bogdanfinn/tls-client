package tls_client_cffi_src

import (
	"bytes"
	"io"
	"testing"

	http "github.com/bogdanfinn/fhttp"
)

func TestBuildResponse_ForcedEncoding_GBK(t *testing.T) {
	gbkBytes := []byte{0xc4, 0xe3, 0xba, 0xc3}
	utf8Expected := "你好"

	encoding := "gbk"
	reqInput := RequestInput{
		RequestMethod: "GET",
		RequestUrl:    "http://example.com",
		ForcedEncoding: &encoding,
	}

	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/html"}},
		Body:       io.NopCloser(bytes.NewReader(gbkBytes)),
	}

	result, err := BuildResponse("", false, resp, nil, reqInput)
	if err != nil {
		t.Fatalf("BuildResponse error: %v", err)
	}

	if result.Body != utf8Expected {
		t.Fatalf("expected body %q, got %q", utf8Expected, result.Body)
	}
}

func TestBuildResponse_ForcedEncoding_Empty(t *testing.T) {
	gbkBytes := []byte{0xc4, 0xe3, 0xba, 0xc3}
	utf8Expected := "你好"

	encoding := ""
	reqInput := RequestInput{
		RequestMethod: "GET",
		RequestUrl:    "http://example.com",
		ForcedEncoding: &encoding,
	}

	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/html; charset=gbk"}},
		Body:       io.NopCloser(bytes.NewReader(gbkBytes)),
	}

	result, err := BuildResponse("", false, resp, nil, reqInput)
	if err != nil {
		t.Fatalf("BuildResponse error: %v", err)
	}

	if result.Body != utf8Expected {
		t.Fatalf("expected body %q, got %q", utf8Expected, result.Body)
	}
}

func TestBuildResponse_AutoDetectFromHeader(t *testing.T) {
	gbkBytes := []byte{0xc4, 0xe3, 0xba, 0xc3}
	utf8Expected := "你好"

	reqInput := RequestInput{
		RequestMethod: "GET",
		RequestUrl:    "http://example.com",
	}

	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/html; charset=gbk"}},
		Body:       io.NopCloser(bytes.NewReader(gbkBytes)),
	}

	result, err := BuildResponse("", false, resp, nil, reqInput)
	if err != nil {
		t.Fatalf("BuildResponse error: %v", err)
	}

	if result.Body != utf8Expected {
		t.Fatalf("expected body %q, got %q", utf8Expected, result.Body)
	}
}

func TestBuildResponse_ForcedEncoding_UTF8(t *testing.T) {
	utf8Bytes := []byte{0xe4, 0xbd, 0xa0, 0xe5, 0xa5, 0xbd}
	utf8Expected := "你好"

	encoding := "utf-8"
	reqInput := RequestInput{
		RequestMethod: "GET",
		RequestUrl:    "http://example.com",
		ForcedEncoding: &encoding,
	}

	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/html"}},
		Body:       io.NopCloser(bytes.NewReader(utf8Bytes)),
	}

	result, err := BuildResponse("", false, resp, nil, reqInput)
	if err != nil {
		t.Fatalf("BuildResponse error: %v", err)
	}

	if result.Body != utf8Expected {
		t.Fatalf("expected body %q, got %q", utf8Expected, result.Body)
	}
}
