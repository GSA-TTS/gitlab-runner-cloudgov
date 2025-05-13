#!/bin/sh

# Move to bundle dir
cd "${0%/*}" || exit 1
dir=$(pwd -P)

. ./glrw-profile.sh

# Sometimes start dir isn't home and home doesn't exist.
if [ "$dir" != "$HOME/bundle" ]; then
    mkdir -p "$HOME"
fi

# started in ./bundle, unpack to home
mv ./* "$HOME"
cd "$HOME" || exit 1

# chmod +x bin/*
tar xf git.tgz
