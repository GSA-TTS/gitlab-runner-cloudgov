#!/usr/bin/env bash
#
# Run a step

currentDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
source "${currentDir}/base.sh"

printf "[cf-driver] Using SSH to connect to %s and run '%s' step\n" "$CONTAINER_ID" "$2"

# Add line below script's shebang to source
# /etc/profile, etc/environment & the $HOME/bin
sed -e '1a\
source /etc/profile\
touch /etc/environment\
source /etc/environment\
source "$HOME/.glr-env"\
PATH="$PATH:/usr/local/bin"\
' "$1" >"$1.tmp"
mv -- "$1.tmp" "$1"

if [ -n "${RUNNER_DEBUG-}" ] && [ "$RUNNER_DEBUG" == "true" ]; then
    # DANGER: There may be sensitive information in this output.
    # Generated job logs should be removed after this is used.
    printf "[cf-driver] RUNNER_DEBUG: About to run the following:\n=========\n"
    cat "$1"
    printf "\n=========\n[cf-driver] RUNNER_DEBUG: End command display\n"
fi

if ! cf_ssh "$CONTAINER_ID" <"${1}"; then
    # Exit using the variable, to make the build as failure in GitLab
    # CI.
    exit "$BUILD_FAILURE_EXIT_CODE"
fi

printf "[cf-driver] Completed SSH session with %s to run '%s' step\n" "$CONTAINER_ID" "$2"
