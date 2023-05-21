#!/bin/bash
# Requires xgo: https://github.com/techknowlogick/xgo

set -e

mkdir dist

echo 'Build OSX'
xgo -buildmode=c-shared -out dist/tls-client --targets=darwin/arm64 .
xgo -buildmode=c-shared -out dist/tls-client --targets=darwin/amd64 .

echo 'Build Linux'
xgo -buildmode=c-shared -out dist/tls-client --targets=linux/arm64 .
xgo -buildmode=c-shared -out dist/tls-client --targets=linux/amd64 .

echo 'Build Windows'
xgo -buildmode=c-shared -out dist/tls-client --targets=windows/386 .
xgo -buildmode=c-shared -out dist/tls-client --targets=windows/amd64 .
