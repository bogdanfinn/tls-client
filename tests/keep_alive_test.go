package tests

import (
	"fmt"
	"io"
	"slices"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/stretchr/testify/assert"
)

func TestClient_UseSameConnection(t *testing.T) {
	testServer := getSimpleWebServer()
	testServer.Start()
	defer testServer.Close()

	client, err := tls_client.ProvideDefaultClient(tls_client.NewNoopLogger())
	if err != nil {
		t.Fatal(err)
	}

	endpoint := fmt.Sprintf("%s%s", testServer.URL, "/index")

	var ports []string
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		responseBody, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		if !slices.Contains(ports, string(responseBody)) {
			ports = append(ports, string(responseBody))
		}

		resp.Body.Close()
	}

	assert.Len(t, ports, 1)
}

func TestClient_UseDifferentConnection(t *testing.T) {
	testServer := getSimpleWebServer()
	testServer.Start()
	defer testServer.Close()

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_107),
		tls_client.WithTransportOptions(&tls_client.TransportOptions{
			DisableKeepAlives: true,
		}),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	endpoint := fmt.Sprintf("%s%s", testServer.URL, "/index")

	var ports []string
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		responseBody, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		if !slices.Contains(ports, string(responseBody)) {
			ports = append(ports, string(responseBody))
		}

		resp.Body.Close()
	}

	assert.Len(t, ports, 5)
}

func getSimpleWebServer() *httptest.Server {
	var indexHandler = func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("receive a request from:", req.RemoteAddr, req.Header)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(req.RemoteAddr))
	}

	router := http.NewServeMux()
	router.HandleFunc("/index", indexHandler)

	ts := httptest.NewUnstartedServer(router)

	return ts
}
