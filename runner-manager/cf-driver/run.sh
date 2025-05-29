#!/usr/bin/env bash
#
# Run a step

currentDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
source "${currentDir}/base.sh"

printf "[cf-driver] Using SSH to connect to %s and run '%s' step\n" "$CONTAINER_ID" "$2"

# If there are exports we go to the end of them
# and then source the profile we created during the prepare step.
#
# Otherwise we add it in a line below script's shebang
#
# shellcheck disable=SC2016 # expands on worker
if grep 'export' "$1" >/dev/null 2>&1; then
    perl -pi -e 's|(.*export.*?\\n)|$1source "\$HOME/glrw-profile.sh"\\n|' "$1"
else
    sed -i '1 a\\nsource "$HOME/glrw-profile.sh"' "$1"
fi

# DANGER: There may be sensitive information in this output.
# Generated job logs should be removed after this is used.
if [ "$RUNNER_DEBUG" == "true" ]; then
    # turn on xtrace after eval so we don't get double output
    # -- this is very similar to CI_DEBUG_TRACE though it may do more
    sed -i "s/eval $'/eval $'set -o xtrace\n/" "$1"

    # Skip cleanup to aid postmortem
    if [ "$2" == "cleanup_file_variables" ]; then
        printf "[cf-driver] RUNNER_DEBUG: skipping cleanup_file_variables"
        exit 0
    fi
fi

cf_ssh "$CONTAINER_ID" <"${1}"
exit_code=$?

# We use SYSTEM_FAILURE_EXIT_CODE here instead of
# BUILD_FAILURE_EXIT_CODE because the former allows retries
# see: https://docs.gitlab.com/runner/executors/custom/#system-failure
#
# The BUILD_EXIT_CODE_FILE facilitates "allow_failure"
# see: https://docs.gitlab.com/runner/executors/custom/#build-failure-exit-code
if [ $exit_code -ne 0 ]; then
    echo $exit_code >"$BUILD_EXIT_CODE_FILE"
    exit "$SYSTEM_FAILURE_EXIT_CODE"
fi

printf "[cf-driver] Completed SSH session with %s to run '%s' step\n" "$CONTAINER_ID" "$2"
