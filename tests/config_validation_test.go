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

func TestConfigValidation_ServerNameOverwriteWithoutInsecure(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithServerNameOverwrite("example.com"),
		// Missing WithInsecureSkipVerify()
	}

	_, err := tls_client.NewHttpClient(nil, options...)
	if err == nil {
		t.Fatal("Expected error when using server name overwrite without insecure skip verify, but got nil")
	}

	expectedMsg := "server name overwrite requires insecure skip verify"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}

	t.Logf("✓ Correctly rejected config with error: %v", err)
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
