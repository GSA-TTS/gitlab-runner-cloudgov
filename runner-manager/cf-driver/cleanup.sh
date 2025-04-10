#!/usr/bin/env bash
#
# Cleanup the executor instance

currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "${currentDir}/base.sh"

set -eo pipefail

# trap any error, and mark it as a system failure.
trap 'exit $SYSTEM_FAILURE_EXIT_CODE' ERR

cleanup_service () {
    alias_name="$1"
    container_id="$2"

    # Delete the service app and the associated route(s)
    cf delete -r -f "$container_id"
}

cleanup_services () {
    container_id_base="$1"
    ci_job_services="$2"

    if [ -z "$ci_job_services" ]; then
       echo "[cf-driver] No services defined in ci_job_services - Skipping service cleanup"
       return
    fi

    for l in $(echo "$ci_job_services" | jq -rc '.[]'); do
        # Using jq -er to fail of alias or name are not found
        alias_name=$(echo "$l" | jq -er '.alias | select(.)')
        container_id="${container_id_base}-svc-${alias_name}"

        echo "[cf-driver] Deleting service $alias_name"
        cleanup_service "$alias_name" "$container_id"
    done
}

if [ "$CUSTOM_ENV_PRESERVE_SERVICES" != "true" ]; then
    cleanup_services "$CONTAINER_ID" "$CUSTOM_ENV_CI_JOB_SERVICES"
fi

if [ "$CUSTOM_ENV_PRESERVE_WORKER" != "true" ]; then
    echo "[cf-driver] Deleting executor instance $CONTAINER_ID"
    cf delete -f "$CONTAINER_ID"
fi

echo "[cf-driver] Cleanup completed for $CONTAINER_ID"
