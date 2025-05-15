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
export GIT_SSL_CAINFO="$SSL_CERT_FILE"
export GIT_PROXY_SSL_CAINFO="$SSL_CERT_FILE"

for f in "$HOME"/glrw-profile.d/*; do
    [ -e "$f" ] || break
    . "$f"
done
