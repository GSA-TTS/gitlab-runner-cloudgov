#!/usr/bin/env bash
#
# Run a step

currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source ${currentDir}/base.sh

echo "[cf-driver] Using SSH to connect to $CONTAINER_ID and run steps"

if [ -n "$RUNNER_DEBUG" ] && [ "$RUNNER_DEBUG" = "true" ]; then
    # DANGER: There may be sensitive information in this output.
    # Generated job logs should be removed after this is used.
    echo "[cf-driver] RUNNER_DEBUG: About to run the following:\n--------"
    cat $1
    echo "--------\n[cf-driver] RUNNER_DEBUG: End command display"
fi

cf ssh "$CONTAINER_ID" < "${1}"
if [ $? -ne 0 ]; then
    # Exit using the variable, to make the build as failure in GitLab
    # CI.
    exit "$BUILD_FAILURE_EXIT_CODE"
fi

echo "[cf-driver] Completed SSH session with $CONTAINER_ID to run steps"
