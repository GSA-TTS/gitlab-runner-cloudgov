#!/usr/bin/env bash

set -euo pipefail

# trap any error, and mark it as a system failure.
# Also cleans up TMPVARFILE (set in create_temporary_varfile).
trap 'rm -f "$TMPVARFILE"; exit $SYSTEM_FAILURE_EXIT_CODE' ERR
trap 'rm -f "$TMPVARFILE"' EXIT

# Prepare a runner executor application in CloudFoundry

currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "${currentDir}/base.sh" # Get variables from base.
if [ -z "${WORKER_MEMORY-}" ]; then
    WORKER_MEMORY="512M"
fi

create_temporary_varfile () {
    # A less leak-prone way to share secrets into the worker which will not
    # be able to parse VCAP_CONFIGURATION
    TMPVARFILE=$(mktemp /tmp/gitlab-runner-worker-manifest.XXXXXXXXXX)

    for v in RUNNER_NAME CACHE_TYPE CACHE_S3_SERVER_ADDRESS CACHE_S3_BUCKET_LOCATION CACHE_S3_BUCKET_NAME CACHE_S3_BUCKET_NAME CACHE_S3_ACCESS_KEY CACHE_S3_SECRET_KEY; do
        echo "$v: \"$v\"" >> "$TMPVARFILE"
    done

    echo "[cf-driver] [DEBUG] Added $(wc -l "$TMPVARFILE") lines to $TMPVARFILE"
}

start_container () {
    if cf app --guid "$CONTAINER_ID" >/dev/null 2>/dev/null ; then
        echo '[cf-driver] Found old instance of runner executor, deleting'
        cf delete "$CONTAINER_ID"
    fi

    cf push "$CONTAINER_ID" -f "${currentDir}/worker-manifest.yml" \
       --docker-image "$CUSTOM_ENV_CI_JOB_IMAGE" -m "$WORKER_MEMORY" \
       --vars-file "$TMPVARFILE"
}

start_service () {
    alias_name="$1"
    container_id="$2"
    image_name="$3"
    container_entrypoint="$4"
    container_command="$5"

    if [ -z "$container_id" ] || [ -z "$image_name" ]; then
       echo 'Usage: start_service CONTAINER_ID IMAGE_NAME CONTAINER_ENTRYPOINT CONTAINER_COMMAND'
       exit 1
    fi
    if [ -n "$container_entrypoint" ] || [ -n "$container_command" ]; then
        # TODO - cf push allows use of -c or --start-command but not a separate
        # entrypoint. May need to add logic to gracefully convert entrypoint to
        # a command.
        echo '[cf-driver] container_entrypoint and container_command are not yet supported in services - Sorry!'
        exit 1
    fi

    if cf app --guid "$container_id" >/dev/null 2>/dev/null ; then
        echo '[cf-driver] Found old instance of runner service, deleting'
        cf delete "$container_id"
    fi

    # TODO - Figure out how to handle command and non-global memory definition
    cf push "$container_id" --docker-image "$image_name" -m "$WORKER_MEMORY" \
        --no-route --health-check-type process

    cf map-route "$container_id" apps.internal --hostname "$container_id"
}

allow_access_to_service () {
    source_app="$1"
    destination_service_app="$2"
    current_org=$(echo "$VCAP_APPLICATION" | jq --raw-output ".organization_name")
    current_space=$(echo "$VCAP_APPLICATION" | jq --raw-output ".space_name")

    # TODO NOTE: This is foolish and allows all TCP ports for now.
    # This is limiting and sloppy.
    protocol="tcp"
    ports="20-10000"

    cf add-network-policy "$source_app" \
        --destination-app "$destination_service_app" \
        -o "$current_org" -s "$current_space" \
        --protocol "$protocol" --port "$ports"
}

start_services () {
    container_id_base="$1"
    ci_job_services="$2"

    if [ -z "$ci_job_services" ]; then
       echo "[cf-driver] No services defined in ci_job_services - Skipping service startup"
       return
    fi

    for l in $(echo "$ci_job_services" | jq -rc '.[]'); do
        alias_name=$(echo "$l" | jq -r '.alias | select(.)')
        container_id="${container_id_base}-svc-${alias_name}"
        image_name=$(echo "$l" | jq -r '.name | select(.)')
        container_entrypoint=$(echo "$l" | jq -r '.entrypoint | select(.)')
        container_command=$(echo "$l" | jq -r '.command | select(.)')

        start_service "$alias_name" "$container_id" "$image_name" "$container_entrypoint" "$container_command"
        allow_access_to_service "$container_id_base" "$container_id"
    done
}

install_dependencies () {
    # Build a command to try and install git and git-lfs on common distros.
    # Of course, RedHat/UBI will need more help to add RPM repos with the correct
    # version. TODO - RedHat support
    echo "[cf-driver] Ensuring git, git-lfs, and curl are installed"
    cf ssh "$CONTAINER_ID" -c '(which git && which git-lfs && which curl) || \
                               (which apk && apk add git git-lfs curl) || \
                               (which apt-get && apt-get update && apt-get install -y git git-lfs curl) || \
                               (echo "Required packages missing and I do not know what to do about it" && exit 1)'

    # gitlab-runner-helper includes a limited subset of gitlab-runner functionality
    # plus Git and Git-LFS. https://s3.dualstack.us-east-1.amazonaws.com/gitlab-runner-downloads/latest/index.html
    #
    # Install gitlab-runner-helper binary since we need to manage cache/artifacts.
    # Symlinks gitlab-runner to avoid having to alter more of the executor.
    # TODO: Pin the version and support more arches than X86_64
    echo "[cf-driver] Installing gitlab-runner-helper"
    cf ssh "$CONTAINER_ID" -c 'curl -L --output /usr/bin/gitlab-runner-helper \
                               "https://s3.dualstack.us-east-1.amazonaws.com/gitlab-runner-downloads/latest/binaries/gitlab-runner-helper/gitlab-runner-helper.x86_64"; \
                               chmod +x /usr/bin/gitlab-runner-helper; \
                               ln -s /usr/bin/gitlab-runner-helper /usr/bin/gitlab-runner'
}

echo "[cf-driver] Preparing environment variables for $CONTAINER_ID"
create_temporary_varfile

echo "[cf-driver] Starting $CONTAINER_ID with image $CUSTOM_ENV_CI_JOB_IMAGE"
start_container

echo "[cf-driver] Installing dependencies into $CONTAINER_ID"
install_dependencies

if [ -n "$CUSTOM_ENV_CI_JOB_SERVICES" ]; then
    echo "[cf-driver] Starting services"
    start_services "$CONTAINER_ID" "$CUSTOM_ENV_CI_JOB_SERVICES"
fi

echo "[cf-driver] $CONTAINER_ID preparation complete"
