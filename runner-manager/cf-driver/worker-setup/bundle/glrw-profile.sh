#!/bin/sh
# shellcheck source=./
# shellcheck disable=1091

# egress_proxy set in manifest during prepare
# shellcheck disable=SC2154
export http_proxy="$egress_proxy"
export https_proxy="$egress_proxy"

if [ -r /etc/profile ]; then
    . /etc/profile
fi

if [ -r /etc/environment ]; then
    . /etc/environment
fi

if [ -r "$HOME/.bash_profile" ]; then
    . "$HOME/.bash_profile"
elif [ -r "$HOME/.bash_login" ]; then
    . "$HOME/.bash_login"
elif [ -r "$HOME/.profile" ]; then
    . "$HOME/.profile"
fi

# If the default cert file isn't writable then we haven't
# been able to update it and will fall back to the copied cert.
if [ ! -w "$SSL_CERT_FILE" ]; then
    export SSL_CERT_FILE="$HOME/certs/ca-certificates.crt"
    export NODE_EXTRA_CA_CERTS="$SSL_CERT_FILE"
fi

export PATH="$HOME/bin:/usr/local/bin:/busybox:$PATH"
export GIT_EXEC_PATH="$HOME/libexec/git-core"
export GIT_TEMPLATE_DIR="$HOME/share/git-core/templates"

if command -v git >/dev/null 2>&1; then
    git config --global http.sslCAInfo "$SSL_CERT_FILE"
    git config --global http.proxySSLCAInfo "$SSL_CERT_FILE"

    # These configs are _probably_ meant to be set by GitLab--it tries
    # to set them, and then immediately deletes them. I think they
    # probably have a `defer` func that isn't where it shouldn't be.
    git config --global init.defaultBranch none
    git config --global fetch.recurseSubmodules false
    git config --global credential.interactive never
    git config --global gc.autoDetach false
    git config --global transfer.bundleURI true
    if [ -n "$CI_SERVER_TLS_CA_FILE" ]; then
        git config --global \
            "http.https://gsa.gitlab-dedicated.us.sslCAInfo" \
            "$CI_SERVER_TLS_CA_FILE"
    fi
fi

for f in "$HOME"/glrw-profile.d/*; do
    [ -e "$f" ] || break
    . "$f"
done
