#!/usr/bin/env bash
#
# Cleanup the executor instance

currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "${currentDir}/base.sh"

set -eo pipefail

# trap any error, and mark it as a system failure.
trap 'exit $SYSTEM_FAILURE_EXIT_CODE' ERR

echo "Deleting executor instance $CONTAINER_ID"

cf delete -f "$CONTAINER_ID"

echo "Cleanup completed for $CONTAINER_ID"
