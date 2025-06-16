#!/bin/sh

# Move to bundle dir
cd "${0%/*}" || exit 1
dir=$(pwd -P)

# source expected paths, hedge paths (e.g. /busybox), and simple env
. ./glrw-profile.sh

set -eu

# Sometimes start dir isn't home and home doesn't exist.
if [ "$dir" != "$HOME/bundle" ]; then
    mkdir -p "$HOME"
fi

# started in ./bundle, unpack to home
if [ -d "$HOME/bin" ]; then
    mv ./bin/* "$HOME/bin"
    rmdir ./bin
fi
mv ./* "$HOME"
cd "$HOME" || exit 1

# extract static git package
tar xf git.tgz

for f in "$HOME"/glrw-setup.d/*; do
    [ -e "$f" ] || break
    $f
done
