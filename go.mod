module github.com/bogdanfinn/tls-client

go 1.24.1

require (
	github.com/bdandy/go-socks4 v1.2.3
	github.com/bogdanfinn/fhttp v0.6.7
	github.com/bogdanfinn/quic-go-utls v1.0.8-utls
	github.com/bogdanfinn/utls v1.7.7-barnius
	github.com/bogdanfinn/websocket v1.5.4-barnius
	github.com/google/uuid v1.6.0
	github.com/stretchr/testify v1.11.1
	github.com/tam7t/hpkp v0.0.0-20160821193359-2b70b4024ed5
	golang.org/x/net v0.48.0
	golang.org/x/text v0.32.0
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/bdandy/go-errors v1.2.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// replace github.com/bogdanfinn/utls => ../utls

// replace github.com/bogdanfinn/quic-go-utls => ../quic-go-utls

// replace github.com/bogdanfinn/websocket => ../websocket

// replace github.com/bogdanfinn/fhttp => ../fhttp
