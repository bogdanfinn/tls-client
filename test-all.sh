#!/bin/sh

nvm use 16

echo 'Run go tests'
go test ./tests/

echo 'Run go examples'
go run ./example/main.go

cd ./cffi_dist/example_node

echo 'Run nodejs examples'
node index.js
node index_custom_client.js
node index_image.js

cd ./../cffi_dist/example_python

echo 'Run python examples'
python3 example.py
python3 example_custom_client.py
python3 example_image.py
