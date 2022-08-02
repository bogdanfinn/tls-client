#!/bin/sh

echo 'Build OSX'
GOOS=darwin CGO_ENABLED=1 GOARCH=arm64 go build -buildmode=c-shared -o ./dist/tls-client-darwin-arm64.dylib
GOOS=darwin CGO_ENABLED=1 GOARCH=amd64 go build -buildmode=c-shared -o ./dist/tls-client-darwin-amd64.dylib

# CC is needed when you cross compile from OSX to Linux
echo 'Build Linux'
GOOS=linux CGO_ENABLED=1 GOARCH=amd64 CC="x86_64-linux-musl-gcc" go build -buildmode=c-shared -o ./dist/tls-client-linux-amd64.so

# CC is needed when you cross compile from OSX to Windows
echo 'Build Windows 32 Bit'
GOOS=windows CGO_ENABLED=1 GOARCH=386 CC="i686-w64-mingw32-gcc" go build -buildmode=c-shared -o ./dist/tls-client-windows-32.dll

# CC is needed when you cross compile from OSX to Windows
echo 'Build Windows 64 Bit'
GOOS=windows CGO_ENABLED=1 GOARCH=amd64 CC="x86_64-w64-mingw32-gcc" go build -buildmode=c-shared -o ./dist/tls-client-windows-64.dll