#!/bin/sh
# shellcheck source=./

if command -v yq >/dev/null 2>&1; then
    export http_proxy=$(echo "$VCAP_SERVICES" | yq '.[][] | select(.name == "worker-egress-credentials") | .credentials.http_uri')
    export https_proxy="$http_proxy"
fi
