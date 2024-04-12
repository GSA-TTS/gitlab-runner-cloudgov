#!/usr/bin/env bash
#
# Run a step

currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source ${currentDir}/base.sh

echo "[cf-driver] Using SSH to connect to $CONTAINER_ID and run steps"

cf ssh "$CONTAINER_ID" -c "${1}"
if [ $? -ne 0 ]; then
    # Exit using the variable, to make the build as failure in GitLab
    # CI.
    exit "$BUILD_FAILURE_EXIT_CODE"
fi

echo "[cf-driver] Completed SSH session with $CONTAINER_ID to run steps"
