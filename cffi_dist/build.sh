#!/bin/sh

# This build.sh file was created on a OSX host system. If you are running on windows / unix you need to adjust the commands accordingly.

echo 'Build with xgo'
xgo -buildmode=c-shared -out dist/tls-client-xgo-$1 .