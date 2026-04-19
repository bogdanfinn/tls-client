package tests

import (
	"fmt"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/stretchr/testify/assert"
)

func TestClient_UseCompressedResponse(t *testing.T) {
	testServer := getSimpleWebServerCompressed()
	testServer.Start()
	defer testServer.Close()

	clientOptions := []tls_client.HttpClientOption{
		tls_client.WithTransportOptions(&tls_client.TransportOptions{
			DisableCompression: true,
		}),
	}
	client, err := tls_client.NewHttpClient(nil, clientOptions...)
	if err != nil {
		t.Fatal(err)
	}

	endpoint := fmt.Sprintf("%s%s", testServer.URL, "/index")
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	req.Header.Add("Accept-Encoding", "gzip, deflate, br, zstd")
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, resp.Uncompressed, false)
}

func getSimpleWebServerCompressed() *httptest.Server {
	var indexHandler = func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("receive a request from:", req.RemoteAddr, req.Header)
		w.Write([]byte(req.RemoteAddr))
		w.WriteHeader(http.StatusOK)
	}

	router := http.NewServeMux()
	router.HandleFunc("/index", indexHandler)

	ts := httptest.NewUnstartedServer(router)

	return ts
}
