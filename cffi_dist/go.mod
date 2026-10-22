module tls_client_cffi

go 1.27

require (
	github.com/bogdanfinn/fhttp v0.7.0
	github.com/bogdanfinn/tls-client v1.16.0
	github.com/google/uuid v1.6.0
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/bdandy/go-errors v1.2.2 // indirect
	github.com/bdandy/go-socks4 v1.2.3 // indirect
	github.com/bogdanfinn/quic-go-utls v1.0.11-utls // indirect
	github.com/bogdanfinn/utls v1.8.2-barnius // indirect
	github.com/bogdanfinn/websocket v1.5.7-barnius // indirect
	github.com/klauspost/compress v1.20.1 // indirect
	github.com/molecule-man/go-brrr v1.2.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/tam7t/hpkp v0.0.0-20160821193359-2b70b4024ed5 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
)

// fork: cffi_src/websocket.go is not part of any upstream release, build against the local checkout
replace github.com/bogdanfinn/tls-client => ../
