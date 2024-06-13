#!/usr/bin/env bash
#
# Cleanup the executor instance

currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "${currentDir}/base.sh"

set -eo pipefail

# trap any error, and mark it as a system failure.
trap 'exit $SYSTEM_FAILURE_EXIT_CODE' ERR

cleanup_services () {
    container_id_base="$1"
    ci_job_services="$2"

    if [ -z "$ci_job_services" ]; then
       echo "No services defined in ci_job_services - Skipping service cleanup"
       return
    fi

    for l in $(echo "$ci_job_services" | jq -rc '.[]'); do
        container_id="${container_id_base}-svc-"$(echo "$l" | jq -r '. | .alias | select(. != null)')

        cf delete -f "$container_id"
    done
}

if [ -n "$CUSTOM_ENV_CI_JOB_SERVICES" ]; then
    echo "[cf-driver] Cleaning up services"
    cleanup_services "$CONTAINER_ID" "$CUSTOM_ENV_CI_JOB_SERVICES"
fi

echo "Deleting executor instance $CONTAINER_ID"

cf delete -f "$CONTAINER_ID"

echo "Cleanup completed for $CONTAINER_ID"
