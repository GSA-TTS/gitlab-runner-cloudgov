#!/bin/sh

# Move to bundle dir
cd "${0%/*}" || exit 1
dir=$(pwd -P)

# source expected paths, hedge paths (e.g. /busybox), and simple env
. ./glrw-profile.sh

# Sometimes start dir isn't home and home doesn't exist.
if [ "$dir" != "$HOME/bundle" ]; then
    mkdir -p "$HOME"
fi

# started in ./bundle, unpack to home
mv ./* "$HOME"
cd "$HOME" || exit 1

# extract static git package
tar xf git.tgz

# extract fallback ssl cert bundle
mkdir certs
tar xf certs.tgz --directory certs

ca_dir=/etc/ssl/certs
sys_crts=/etc/cf-system-certificates
custom_dir=/usr/local/share/ca-certificates
update_cmd=update-ca-certificates

# If we have access to the debian/ubuntu custom CA directory
# and the `update-ca-certificates` command, then we can just
# run that and be done.
if [ ! -e "$custom_dir" ]; then
    mkdir -p "$custom_dir" ||
        echo "[glrw-setup] $custom_dir unfound and could not create"
fi
if [ -w "$custom_dir" ] && command -v $update_cmd >/dev/null; then
    cp "$sys_crts"/* "$custom_dir"
    $update_cmd && exit
fi

# Some systems & software will look for a CA directory.
#
# If the standard (at least to debian/ubuntu) directory
# doesn't exist we attempt to create it and populate it with
# our bundled fallback certs.
#
# If it does exist we try to copy the CF system certs into it.
if [ ! -e "$ca_dir" ]; then
    mkdir -p "$ca_dir" ||
        echo "[glrw-setup] $ca_dir unfound and could not create"

    cp ./certs/* "$ca_dir" ||
        echo "[glrw-setup] $ca_dir created but failed to populate"
else
    cp "$sys_crts"/* $ca_dir ||
        echo "[glrw-setup] $ca_dir found but failed to copy system certs"
fi

# Some systems will look for a single CA bundle, this seems
# more common so far.
#
# At least debian/ubuntu will check for certs/ca-certificates.crt,
# some systems & software will look for cert.pem.
#
# We try to update the existing bundle if it exists,
# otherwise we try to link it to our falback cert bundle.
ca_bundles="/etc/ssl/certs/ca-certificates.crt /etc/ssl/cert.pem"
for bundle in $ca_bundles; do
    if [ ! -w "$(dirname "$bundle")" ] && [ ! -w "$bundle" ]; then
        echo "[glrw-setup] $bundle not writable"
    elif [ -e "$bundle" ]; then
        cat "$sys_crts"/* >>"$bundle" ||
            echo "[glrw-setup] could not append to $bundle"
    else
        ln -s "$HOME/certs/ca-certificates.crt" "$bundle" ||
            echo "[glrw-setup] could create link for $bundle"
    fi
done
