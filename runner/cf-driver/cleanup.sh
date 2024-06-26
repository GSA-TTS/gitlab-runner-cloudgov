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

remove_access_to_service () {
    source_app="$1"
    destination_service_app="$2"
    current_org=$(echo "$VCAP_APPLICATION" | jq --raw-output ".organization_name")
    current_space=$(echo "$VCAP_APPLICATION" | jq --raw-output ".space_name")

    # TODO NOTE: This is foolish and allows all TCP ports for now.
    # This is limiting and sloppy.
    protocol="tcp"
    ports="20-10000"

    cf remove-network-policy "$source_app" \
        --destination-app "$destination_service_app" \
        -o "$current_org" -s "$current_space" \
        --protocol "$protocol" --port "$ports"
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

        echo "[cf-driver] Removing network policy from $container_id_base to $container_id"
        remove_access_to_service "$container_id_base" "$container_id"

        echo "[cf-driver] Deleting service $alias_name"
        cleanup_service "$alias_name" "$container_id"
    done
}

cleanup_services "$CONTAINER_ID" "$CUSTOM_ENV_CI_JOB_SERVICES"

echo "[cf-driver] Deleting executor instance $CONTAINER_ID"
cf delete -f "$CONTAINER_ID"

echo "[cf-driver] Cleanup completed for $CONTAINER_ID"
