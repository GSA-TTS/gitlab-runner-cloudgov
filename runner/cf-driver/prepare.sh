#!/usr/bin/env bash

set -euo pipefail

# trap any error, and mark it as a system failure.

# Also cleans up TMPMANIFEST(set in create_temporary_manifest).
trap 'rm -f "$TMPMANIFEST"; exit $SYSTEM_FAILURE_EXIT_CODE' ERR
trap 'rm -f "$TMPMANIFEST"' EXIT

# Prepare a runner executor application in CloudFoundry

currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "${currentDir}/base.sh" # Get variables from base.
if [ -z "${WORKER_MEMORY-}" ]; then
    # Some jobs may fail with less than 512M, e.g., `npm i`
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

get_registry_credentials () {
    image_name="$1"

    # Note: the regex for non-docker image locations is not air-tight--
    #       the definition for the format is a little loose, for one thing,
    #       but this should work for most cases and can be revisited when
    #       we're working with more a more robust set of language features
    #       and can better parse the image name.

    if echo "$image_name" | grep -q "registry.gitlab.com"; then
        # Detect GitLab CR and use provided environment to authenticate
        echo "$CUSTOM_ENV_CI_REGISTRY_USER" "$CUSTOM_ENV_CI_REGISTRY_PASSWORD"

    elif echo "$image_name" | grep -q -P '^(?!registry-\d+.docker.io)[\w-]+(?:\.[\w-]+)+'; then
        # Detect non-Docker registry that we aren't supporting auth for yet
        return 0

    elif [ -n "$DOCKER_HUB_TOKEN" ] && [ -n "$DOCKER_HUB_USER" ]; then
        # Default to Docker Hub credentials when available
        echo "$DOCKER_HUB_USER" "$DOCKER_HUB_TOKEN"
    fi
}

create_temporary_manifest () {
    # A less leak-prone way to share secrets into the worker which will not
    # be able to parse VCAP_CONFIGURATION
    TMPMANIFEST=$(mktemp /tmp/gitlab-runner-worker-manifest.XXXXXXXXXX)
    chmod 600 "$TMPMANIFEST"
    cat "${currentDir}/worker-manifest.yml" > "$TMPMANIFEST"

    # Align additional environment variables with YAML at end of source manifest
    local padding="      "

    for v in RUNNER_NAME CACHE_TYPE CACHE_S3_SERVER_ADDRESS CACHE_S3_BUCKET_LOCATION CACHE_S3_BUCKET_NAME CACHE_S3_ACCESS_KEY CACHE_S3_SECRET_KEY; do
        echo "${padding}${v}: \"${!v}\"" >> "$TMPMANIFEST"
    done

    # Add any CI_SERVICE_x variables populated by start_service()
    for v in "${!CI_SERVICE_@}"; do
        echo "${padding}${v}: \"${!v}\"" >> "$TMPMANIFEST"
    done

    echo "[cf-driver] [DEBUG] $(wc -l < "$TMPMANIFEST") lines in $TMPMANIFEST"
}

start_container () {
    container_id="$1"
    image_name="$CUSTOM_ENV_CI_JOB_IMAGE"

    if cf app --guid "$container_id" >/dev/null 2>/dev/null ; then
        echo '[cf-driver] Found old instance of runner executor, deleting'
        cf delete -f "$container_id"

    fi

    push_args=(
       "$container_id"
       -f "$TMPMANIFEST"
       -m "$WORKER_MEMORY"
       --docker-image "$image_name"
    )

    local docker_user docker_pass
    read -r docker_user docker_pass <<< "$(get_registry_credentials "$image_name")"

    if [ -n "$docker_user" ] && [ -n "$docker_pass" ]; then
        push_args+=('--docker-username' "${docker_user}")
        local -x CF_DOCKER_PASSWORD="${docker_pass}"
    fi

    cf push "${push_args[@]}"
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

    if cf app --guid "$container_id" >/dev/null 2>/dev/null ; then
        echo '[cf-driver] Found old instance of runner service, deleting'
        cf delete -f "$container_id"
    fi

    push_args=(
        "$container_id"
        '-m' "$WORKER_MEMORY"
        '--docker-image' "$image_name"
        '--health-check-type' 'process'
        '--no-route'
    )

    if [ -n "$container_entrypoint" ] || [ -n "$container_command" ]; then
        push_args+=('-c' "${container_entrypoint[@]}" "${container_command[@]}")
    fi

    local docker_user docker_pass
    read -r docker_user docker_pass <<< "$(get_registry_credentials "$image_name")"

    if [ -n "$docker_user" ] && [ -n "$docker_pass" ]; then
        push_args+=('--docker-username' "${docker_user}")
        local -x CF_DOCKER_PASSWORD="${docker_pass}"
    fi

    # TODO - Figure out how to handle non-global memory definition
    cf push "${push_args[@]}"

    # Map route and export a FQDN. We assume apps.internal as the domain.
    cf map-route "$container_id" apps.internal --hostname "$container_id"
    export "CI_SERVICE_${alias_name}"="${container_id}.apps.internal"
}

start_services () {
    container_id_base="$1"
    ci_job_services="$2"

    if [ -z "$ci_job_services" ]; then
       echo "[cf-driver] No services defined in ci_job_services - Skipping service startup"
       return
    fi

    for l in $(echo "$ci_job_services" | jq -rc '.[]'); do
        # Using jq -er to fail of alias or name are not found
        alias_name=$(echo "$l" | jq -er '.alias | select(.)')
        container_id="${container_id_base}-svc-${alias_name}"
        image_name=$(echo "$l" | jq -er '.name | select(.)')
        # Using jq -r to allow entrypoint and command to be empty
        container_entrypoint=$(echo "$l" | jq -r '.entrypoint | select(.)')
        container_command=$(echo "$l" | jq -r '.command | select(.)')

        start_service "$alias_name" "$container_id" "$image_name" "$container_entrypoint" "$container_command"
    done
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


allow_access_to_services () {
    container_id_base="$1"
    ci_job_services="$2"

    if [ -z "$ci_job_services" ]; then
       echo "[cf-driver] No services defined in ci_job_services - Skipping service allowance"
       return
    fi

    for l in $(echo "$ci_job_services" | jq -rc '.[]'); do
        container_id="${container_id_base}-svc-${alias_name}"
        allow_access_to_service "$container_id_base" "$container_id"
    done
}

install_dependencies () {
    container_id="$1"

    # Build a command to try and install git and git-lfs on common distros.
    # Of course, RedHat/UBI will need more help to add RPM repos with the correct
    # version. TODO - RedHat support
    echo "[cf-driver] Ensuring git, git-lfs, and curl are installed"
    cf ssh "$container_id" \
        --command 'source /etc/profile && (which git && which git-lfs && which curl) || \
                               (which apk && apk add git git-lfs curl) || \
                               (which apt-get && apt-get update && apt-get install -y git git-lfs curl) || \
                               (which yum && yum install git git-lfs curl) || \
                               (echo "[cf-driver] Required packages missing and install attempt failed" && exit 1)'

    # gitlab-runner-helper includes a limited subset of gitlab-runner functionality
    # plus Git and Git-LFS. https://s3.dualstack.us-east-1.amazonaws.com/gitlab-runner-downloads/latest/index.html
    #
    # Install gitlab-runner-helper binary since we need to manage cache/artifacts.
    # Symlinks gitlab-runner to avoid having to alter more of the executor.
    # TODO: Pin the version and support more arches than X86_64
    echo "[cf-driver] Installing gitlab-runner-helper"
    cf ssh "$container_id" -c 'curl -L --output /usr/bin/gitlab-runner-helper \
                               "https://s3.dualstack.us-east-1.amazonaws.com/gitlab-runner-downloads/latest/binaries/gitlab-runner-helper/gitlab-runner-helper.x86_64"; \
                               chmod +x /usr/bin/gitlab-runner-helper; \
                               ln -s /usr/bin/gitlab-runner-helper /usr/bin/gitlab-runner'
}

if [ -n "$CUSTOM_ENV_CI_JOB_SERVICES" ]; then
    echo "[cf-driver] Starting services"
    start_services "$CONTAINER_ID" "$CUSTOM_ENV_CI_JOB_SERVICES"
fi

echo "[cf-driver] Preparing manifest for $CONTAINER_ID"
create_temporary_manifest

echo "[cf-driver] Starting $CONTAINER_ID with image $CUSTOM_ENV_CI_JOB_IMAGE"
start_container "$CONTAINER_ID"

echo "[cf-driver] Installing dependencies into $CONTAINER_ID"
install_dependencies "$CONTAINER_ID"

# Allowing access last so services and the worker are all present
if [ -n "$CUSTOM_ENV_CI_JOB_SERVICES" ]; then
    echo "[cf-driver] Enabling access from worker to services"
    allow_access_to_services "$CONTAINER_ID" "$CUSTOM_ENV_CI_JOB_SERVICES"
fi

echo "[cf-driver] $CONTAINER_ID preparation complete"
