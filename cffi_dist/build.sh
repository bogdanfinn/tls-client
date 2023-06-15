#!/bin/sh

# This build.sh file was created on a OSX host system. If you are running on windows / unix you need to adjust the commands accordingly.

echo 'Build OSX'
GOOS=darwin CGO_ENABLED=1 GOARCH=arm64 go build -buildmode=c-shared -o ./dist/tls-client-darwin-arm64-ventura.dylib
GOOS=darwin CGO_ENABLED=1 GOARCH=amd64 go build -buildmode=c-shared -o ./dist/tls-client-darwin-amd64-ventura.dylib
#GOOS=darwin CGO_ENABLED=1 GOARCH=arm64 go build -buildmode=c-shared -o ./dist/tls-client-darwin-arm64-.dylib
#GOOS=darwin CGO_ENABLED=1 GOARCH=amd64 go build -buildmode=c-shared -o ./dist/tls-client-darwin-amd64-.dylib

# CC is needed when you cross compile from OSX to Linux
echo 'Build Linux Ubuntu'
# For some reason my OSX gcc cross compiler does not work. Therefore i use a ubuntu docker image
# GOOS=linux CGO_ENABLED=1 GOARCH=amd64 CC="x86_64-linux-musl-gcc" go build -buildmode=c-shared -o ./dist/tls-client-linux-amd64.so
# Make sure to first build the image based on the Dockerfile.ubuntu.compile in this directory.
# docker build . -t tls-client tls-client-ubuntu-go-1.18
docker run --platform linux/x86_64 -v $PWD/../:/tls-client tls-client-ubuntu-go-1.18 bash -c "cd /tls-client/cffi_dist && GOOS=linux CGO_ENABLED=1 GOARCH=amd64 go build -buildmode=c-shared -o /tls-client/cffi_dist/dist/tls-client-linux-ubuntu-amd64.so"