#!/bin/bash

set -o errexit -o nounset -o pipefail

echo "Get the repo for ICS721 proxy contracts"

if [ ! -d "cw-ics721-proxy" ]; then
	git clone https://github.com/gjermundgaraba/cw-ics721-proxy.git
fi

cd cw-ics721-proxy
git checkout gjermund/incoming-proxy-whitelist
./build.sh
cp ./artifacts/cw_ics721_incoming_proxy_base.wasm ../interslothtest/test-artifacts/cw_ics721_incoming_proxy_base.wasm