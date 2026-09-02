package tests

import (
	"net"
	"strings"
	"testing"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"golang.org/x/net/proxy"
)

func TestConfigValidation_HTTP3RacingWithDisableHTTP3(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithProtocolRacing(),
		tls_client.WithDisableHttp3(),
	}

	_, err := tls_client.NewHttpClient(nil, options...)
	if err == nil {
		t.Fatal("Expected error when enabling HTTP/3 racing with HTTP/3 disabled, but got nil")
	}

	expectedMsg := "HTTP/3 racing cannot be enabled when HTTP/3 is disabled"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}

	t.Logf("✓ Correctly rejected config with error: %v", err)
}

func TestConfigValidation_HTTP3RacingWithForceHTTP1(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithProtocolRacing(),
		tls_client.WithForceHttp1(),
	}

	_, err := tls_client.NewHttpClient(nil, options...)
	if err == nil {
		t.Fatal("Expected error when enabling HTTP/3 racing with HTTP/1 forced, but got nil")
	}

	expectedMsg := "HTTP/3 racing cannot be enabled when HTTP/1 is forced"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}

	t.Logf("✓ Correctly rejected config with error: %v", err)
}

func TestConfigValidation_DisableBothIPVersions(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithDisableIPV4(),
		tls_client.WithDisableIPV6(),
	}

	_, err := tls_client.NewHttpClient(nil, options...)
	if err == nil {
		t.Fatal("Expected error when disabling both IPv4 and IPv6, but got nil")
	}

	expectedMsg := "cannot disable both IPv4 and IPv6"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}

	t.Logf("✓ Correctly rejected config with error: %v", err)
}

func TestConfigValidation_ValidConfigs(t *testing.T) {
	testCases := []struct {
		name    string
		options []tls_client.HttpClientOption
	}{
		{
			name: "HTTP/3 racing enabled (default)",
			options: []tls_client.HttpClientOption{
				tls_client.WithClientProfile(profiles.Chrome_133),
				tls_client.WithProtocolRacing(),
			},
		},
		{
			name: "HTTP/3 racing with IPv4 only",
			options: []tls_client.HttpClientOption{
				tls_client.WithClientProfile(profiles.Chrome_133),
				tls_client.WithProtocolRacing(),
				tls_client.WithDisableIPV6(),
			},
		},
		{
			name: "HTTP/3 racing with IPv6 only",
			options: []tls_client.HttpClientOption{
				tls_client.WithClientProfile(profiles.Chrome_133),
				tls_client.WithProtocolRacing(),
				tls_client.WithDisableIPV4(),
			},
		},
		{
			name: "Force HTTP/1 with HTTP/3 disabled",
			options: []tls_client.HttpClientOption{
				tls_client.WithClientProfile(profiles.Chrome_133),
				tls_client.WithForceHttp1(),
				tls_client.WithDisableHttp3(),
			},
		},
		{
			name: "Disable HTTP/3 without racing",
			options: []tls_client.HttpClientOption{
				tls_client.WithClientProfile(profiles.Chrome_133),
				tls_client.WithDisableHttp3(),
			},
		},
		{
			name: "Force HTTP/1 without racing",
			options: []tls_client.HttpClientOption{
				tls_client.WithClientProfile(profiles.Chrome_133),
				tls_client.WithForceHttp1(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := tls_client.NewHttpClient(nil, tc.options...)
			if err != nil {
				t.Fatalf("Expected valid config to be accepted, but got error: %v", err)
			}

			if client == nil {
				t.Fatal("Expected client to be created, but got nil")
			}

			t.Logf("✓ Config accepted: %s", tc.name)
		})
	}
}

func TestConfigValidation_CertificatePinningWithInsecureSkipVerify(t *testing.T) {
	pins := map[string][]string{
		"example.com": {"pin1", "pin2"},
	}

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithCertificatePinning(pins, nil),
		tls_client.WithInsecureSkipVerify(),
	}

	_, err := tls_client.NewHttpClient(nil, options...)
	if err == nil {
		t.Fatal("Expected error when using certificate pinning with insecure skip verify, but got nil")
	}

	expectedMsg := "certificate pinning cannot be used with insecure skip verify"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}

	t.Logf("✓ Correctly rejected config with error: %v", err)
}

func TestConfigValidation_ProxyUrlAndDialerFactory(t *testing.T) {
	customDialerFactory := func(proxyUrlStr string, timeout time.Duration, localAddr *net.TCPAddr, connectHeaders http.Header, logger tls_client.Logger) (proxy.ContextDialer, error) {
		return proxy.Direct, nil
	}

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithProxyUrl("http://proxy.example.com:8080"),
		tls_client.WithProxyDialerFactory(customDialerFactory),
	}

	_, err := tls_client.NewHttpClient(nil, options...)
	if err == nil {
		t.Fatal("Expected error when setting both proxy URL and custom dialer factory, but got nil")
	}

	expectedMsg := "cannot set both proxy URL and custom proxy dialer factory"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}

	t.Logf("✓ Correctly rejected config with error: %v", err)
}

func TestConfigValidation_ServerNameOverwriteWithInsecure(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithServerNameOverwrite("example.com"),
		tls_client.WithInsecureSkipVerify(),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatalf("Expected valid config to be accepted, but got error: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be created, but got nil")
	}

	t.Log("✓ Server name overwrite with insecure skip verify is correctly accepted")
}

func TestConfigValidation_CertificatePinningWithoutInsecure(t *testing.T) {
	pins := map[string][]string{
		"example.com": {"pin1", "pin2"},
	}

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithCertificatePinning(pins, nil),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatalf("Expected valid config to be accepted, but got error: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be created, but got nil")
	}

	t.Log("✓ Certificate pinning without insecure skip verify is correctly accepted")
}

func TestConfigValidation_OrderIndependent(t *testing.T) {
	t.Run("Racing first, then disable HTTP/3", func(t *testing.T) {
		options := []tls_client.HttpClientOption{
			tls_client.WithProtocolRacing(),
			tls_client.WithDisableHttp3(),
			tls_client.WithClientProfile(profiles.Chrome_133),
		}

		_, err := tls_client.NewHttpClient(nil, options...)
		if err == nil {
			t.Fatal("Expected error regardless of option order")
		}
	})

	t.Run("Disable HTTP/3 first, then racing", func(t *testing.T) {
		options := []tls_client.HttpClientOption{
			tls_client.WithDisableHttp3(),
			tls_client.WithProtocolRacing(),
			tls_client.WithClientProfile(profiles.Chrome_133),
		}

		_, err := tls_client.NewHttpClient(nil, options...)
		if err == nil {
			t.Fatal("Expected error regardless of option order")
		}
	})

	t.Run("IPv4 disable first", func(t *testing.T) {
		options := []tls_client.HttpClientOption{
			tls_client.WithDisableIPV4(),
			tls_client.WithDisableIPV6(),
			tls_client.WithClientProfile(profiles.Chrome_133),
		}

		_, err := tls_client.NewHttpClient(nil, options...)
		if err == nil {
			t.Fatal("Expected error regardless of option order")
		}
	})

	t.Run("IPv6 disable first", func(t *testing.T) {
		options := []tls_client.HttpClientOption{
			tls_client.WithDisableIPV6(),
			tls_client.WithDisableIPV4(),
			tls_client.WithClientProfile(profiles.Chrome_133),
		}

		_, err := tls_client.NewHttpClient(nil, options...)
		if err == nil {
			t.Fatal("Expected error regardless of option order")
		}
	})
}

// --- Proxy + HTTP/3 validation tests ---

func TestConfigValidation_ProtocolRacingWithHTTPProxy(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithProtocolRacing(),
		tls_client.WithProxyUrl("http://proxy.example.com:8080"),
	}

	_, err := tls_client.NewHttpClient(nil, options...)
	if err == nil {
		t.Fatal("Expected error when enabling protocol racing with HTTP proxy")
	}

	expectedMsg := "protocol racing requires a SOCKS5 proxy"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}

	t.Logf("✓ Correctly rejected HTTP proxy with racing: %v", err)
}

func TestConfigValidation_ProtocolRacingWithHTTPSProxy(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithProtocolRacing(),
		tls_client.WithProxyUrl("https://proxy.example.com:443"),
	}

	_, err := tls_client.NewHttpClient(nil, options...)
	if err == nil {
		t.Fatal("Expected error when enabling protocol racing with HTTPS proxy")
	}

	expectedMsg := "protocol racing requires a SOCKS5 proxy"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}

	t.Logf("✓ Correctly rejected HTTPS proxy with racing: %v", err)
}

func TestConfigValidation_ProtocolRacingWithSOCKS4Proxy(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithProtocolRacing(),
		tls_client.WithProxyUrl("socks4://proxy.example.com:1080"),
	}

	_, err := tls_client.NewHttpClient(nil, options...)
	if err == nil {
		t.Fatal("Expected error when enabling protocol racing with SOCKS4 proxy")
	}

	expectedMsg := "protocol racing requires a SOCKS5 proxy"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}

	t.Logf("✓ Correctly rejected SOCKS4 proxy with racing: %v", err)
}

func TestConfigValidation_ProtocolRacingWithSOCKS5Proxy(t *testing.T) {
	// SOCKS5 + racing should be accepted (no actual proxy needed for config validation)
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithProtocolRacing(),
		tls_client.WithProxyUrl("socks5://127.0.0.1:1080"),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatalf("Expected SOCKS5 proxy with racing to be accepted, got error: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to be created")
	}

	t.Log("✓ SOCKS5 proxy with protocol racing accepted")
}

func TestConfigValidation_ProtocolRacingWithSOCKS5hProxy(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithProtocolRacing(),
		tls_client.WithProxyUrl("socks5h://user:pass@127.0.0.1:1080"),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatalf("Expected SOCKS5h proxy with racing to be accepted, got error: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to be created")
	}

	t.Log("✓ SOCKS5h proxy with protocol racing accepted")
}

func TestConfigValidation_HTTPProxyWithoutRacing(t *testing.T) {
	// HTTP proxy without racing should be fine (HTTP/3 won't be attempted)
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithProxyUrl("http://proxy.example.com:8080"),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatalf("Expected HTTP proxy without racing to be accepted, got error: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to be created")
	}

	t.Log("✓ HTTP proxy without racing accepted")
}

func TestConfigValidation_NoProxyWithRacing(t *testing.T) {
	// Racing without proxy should be fine
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithProtocolRacing(),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatalf("Expected racing without proxy to be accepted, got error: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to be created")
	}

	t.Log("✓ Racing without proxy accepted")
}

// The proxy checks in validateConfig only run when the client is built, so SetProxy has
// to repeat them. Otherwise a client created without a proxy could be switched to a TCP
// only proxy at runtime and let the HTTP/3 leg of the race bypass it.
func TestConfigValidation_SetProxyRevalidatesProtocolRacing(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithProtocolRacing(),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatalf("Expected client without proxy to be created, got error: %v", err)
	}

	err = client.SetProxy("http://proxy.example.com:8080")
	if err == nil {
		t.Fatal("Expected SetProxy to reject an HTTP proxy while protocol racing is enabled")
	}

	expectedMsg := "protocol racing requires a SOCKS5 proxy"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}

	if got := client.GetProxy(); got != "" {
		t.Fatalf("Expected the rejected proxy not to be applied, GetProxy() = %q", got)
	}

	if err := client.SetProxy("socks5://127.0.0.1:1080"); err != nil {
		t.Fatalf("Expected SetProxy to accept a SOCKS5 proxy, got error: %v", err)
	}

	t.Logf("✓ SetProxy revalidates the racing/proxy combination")
}
