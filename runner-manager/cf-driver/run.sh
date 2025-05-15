#!/usr/bin/env bash
#
# Run a step

currentDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
source "${currentDir}/base.sh"

printf "[cf-driver] Using SSH to connect to %s and run '%s' step\n" "$CONTAINER_ID" "$2"

# Add line below script's shebang to source
# the profile we created during the prepare step.
# shellcheck disable=2016
sed -e '1a\
source "$HOME/glrw-profile.sh"\
' "$1" >"$1.tmp"
mv -- "$1.tmp" "$1"

if [ -n "${RUNNER_DEBUG-}" ] && [ "$RUNNER_DEBUG" == "true" ]; then
    if [ "$2" == "cleanup_file_variables" ]; then
        printf "[cf-driver] RUNNER_DEBUG: skipping cleanup_file_variables"
        exit 0
    fi

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
