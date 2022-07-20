module github.com/bogdanfinn/tls-client

go 1.18

require (
	github.com/Carcraftz/fhttp v0.0.0-00010101000000-000000000000
	github.com/Carcraftz/utls v0.0.0-20220413235215-6b7c52fd78b6
	golang.org/x/net v0.0.0-20220622184535-263ec571b305
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	gitlab.com/yawning/bsaes.git v0.0.0-20190805113838-0a714cd429ec // indirect
	gitlab.com/yawning/utls.git v0.0.12-1 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/text v0.3.7 // indirect
)

//replace github.com/Carcraftz/utls => ../utls
replace github.com/Carcraftz/utls => github.com/bogdanfinn/utls v0.0.2

//replace github.com/Carcraftz/fhttp => ../fhttp
replace github.com/Carcraftz/fhttp => github.com/bogdanfinn/fhttp v0.0.2
