#!/usr/bin/env bash

set -euo pipefail

# trap any error, and mark it as a system failure.
# Also cleans up TMPFILES created in create_temporary_manifest and start_service
trap 'rm -f "${TMPFILES[@]}"; exit $SYSTEM_FAILURE_EXIT_CODE' ERR
trap 'rm -f "${TMPFILES[@]}"' EXIT
TMPFILES=()

# Prepare a runner executor application in CloudFoundry

currentDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
source "${currentDir}/base.sh" # Get variables from base.
if [ -z "${WORKER_MEMORY-}" ]; then
    # Some jobs may fail with less than 512M, e.g., `npm i`
    WORKER_MEMORY="768M"
fi

get_registry_credentials() {
    image_name="$1"

    # Note: the regex for non-docker image locations is not air-tight--
    #       the definition for the format is a little loose, for one thing,
    #       but this should work for most cases and can be revisited when
    #       we're working with more a more robust set of language features
    #       and can better parse the image name.

    if echo "$image_name" | grep -q "$CUSTOM_ENV_CI_REGISTRY"; then
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

create_temporary_manifest() {
    # A less leak-prone way to share secrets into the worker which will not
    # be able to parse VCAP_CONFIGURATION
    TMPMANIFEST=$(mktemp /tmp/gitlab-runner-worker-manifest.XXXXXXXXXX)
    TMPFILES+=("$TMPMANIFEST")
    chmod 600 "$TMPMANIFEST"
    cat "${currentDir}/worker-manifest.yml" >"$TMPMANIFEST"

    # Align additional environment variables with YAML at end of source manifest
    local padding="      "

    # Add any CI_SERVICE_x variables populated by start_service()
    for v in "${!CI_SERVICE_@}"; do
        echo "${padding}${v}: \"${!v}\"" >>"$TMPMANIFEST"
    done
    for v in "${!CI_SERVICE_ID_@}"; do
        echo "${padding}${v}: \"${!v}\"" >>"$TMPMANIFEST"
    done

    echo "[cf-driver] [DEBUG] $(wc -l <"$TMPMANIFEST") lines in $TMPMANIFEST"
}

setup_proxy_access() {
    container_id="$1"

    # setup network policy to egress-proxy
    cf add-network-policy "$container_id" "$PROXY_APP_NAME" -s "$PROXY_SPACE" \
        --protocol "tcp" --port "61443"
    cf add-network-policy "$container_id" "$PROXY_APP_NAME" -s "$PROXY_SPACE" \
        --protocol "tcp" --port "8080"

    # set environment variables and restart container to pick them up
    cf set-env "$container_id" https_proxy "$http_proxy"
    cf set-env "$container_id" http_proxy "$http_proxy"
    cf restart "$container_id"

    # update ssl certs
    cf_ssh "$container_id" \
        'source /etc/profile && \
        mkdir -p /usr/local/share/ca-certificates && \
        cat /etc/cf-system-certificates/* > /usr/local/share/ca-certificates/cf-system-certificates.crt && \
        (command -v update-ca-certificates && update-ca-certificates) || \
        ( \
            [ -f /etc/ssl/certs/ca-certificates.crt ] && \
            cat /etc/cf-system-certificates/* >> /etc/ssl/certs/ca-certificates.crt \
        ) || \
        ( \
            [ -f /etc/ssl/cert.pem ] && \
            cat /etc/cf-system-certificates/* >> /etc/ssl/cert.pem && \
            ln -s /etc/ssl/cert.pem /etc/ssl/certs/ca-certificates.crt \
        ) || \
        (echo "[cf-driver] Could not update system ca certificates. This may or may not be a problem depending on your base image." && exit 0)'
}

get_start_command() {
    local -n args=$1
    entry="$2"
    command="$3"
    sh_fallback="${4:-false}" # some services may fail w/ sh fallback

    if [ -z "$entry" ] && [ -z "$command" ]; then
        $sh_fallback && args+=('-c' '/bin/sh')
        return 0
    fi

    args+=('-c')
    if [ -n "$entry" ]; then
        args+=("${entry[@]}")
    fi
    if [ -n "$command" ]; then
        args+=("${command[@]}")
    fi
}

start_container() {
    container_id="$1"
    image_name="$CUSTOM_ENV_CI_JOB_IMAGE"

    if cf app --guid "$container_id" >/dev/null 2>/dev/null; then
        echo '[cf-driver] Found old instance of runner executor, deleting'
        cf delete -f "$container_id"
    fi

    local worker_memory=$(jq -r '.variables[]? | select(.key == "WORKER_MEMORY") | .value' "$JOB_RESPONSE_FILE")
    if [ -z "$worker_memory" ]; then
        worker_memory=$WORKER_MEMORY
    fi
    local worker_disk=$(jq -r '.variables[]? | select(.key == "WORKER_DISK") | .value' "$JOB_RESPONSE_FILE")
    if [ -z "$worker_disk" ]; then
        worker_disk=$WORKER_DISK_SIZE
    fi
    local push_args=(
        "$container_id"
        -f "$TMPMANIFEST"
        -m "$worker_memory"
        -k "$worker_disk"
        --docker-image "$image_name"
        --var "cache_bucket=$CACHE_S3_BUCKET_NAME"
    )

    # Entrypoint & command aren't available w/o loading job res file
    img_data=$(jq -rc '.image' "$JOB_RESPONSE_FILE")
    container_entrypoint=$(echo "$img_data" | jq -r '.entrypoint | select(.)')
    container_command=$(echo "$img_data" | jq -r '.command | select(.)')
    get_start_command push_args "${container_entrypoint[@]}" "${container_command[@]}" true

    local docker_user docker_pass
    read -r docker_user docker_pass <<<"$(get_registry_credentials "$image_name")"

    if [ -z "${RUNNER_DEBUG-}" ] || [ "$RUNNER_DEBUG" != "true" ]; then
        push_args+=('--redact-env')
    fi

    if [ -n "$docker_user" ] && [ -n "$docker_pass" ]; then
        push_args+=('--docker-username' "${docker_user}")
        local -x CF_DOCKER_PASSWORD="${docker_pass}"
    fi

    cf push "${push_args[@]}"

    # this must be the very first step after `cf push` as it performs a
    # container restart which will wipe out any other changes made via `cf_ssh`
    echo "[cf-driver] Setting up egress proxy access for $CONTAINER_ID"
    setup_proxy_access "$CONTAINER_ID"
}

start_service() {
    alias_name="$1"
    container_id="$2"
    image_name="$3"
    service_entrypoint="$4"
    service_command="$5"
    service_vars="$6"
    job_vars="$7"

    if [ -z "$container_id" ] || [ -z "$image_name" ]; then
        echo 'Usage: start_service CONTAINER_ID IMAGE_NAME \
            SERVICE_ENTRYPOINT SERVICE_COMMAND \
            SERVICE_VARS JOB_VARS'
        exit 1
    fi

    if cf app --guid "$container_id" >/dev/null 2>/dev/null; then
        echo '[cf-driver] Found old instance of runner service, deleting'
        cf delete -f "$container_id"
    fi

    local push_args=(
        "$container_id"
        '-m' "$WORKER_MEMORY"
        '--docker-image' "$image_name"
        '--health-check-type' 'process'
        '--no-route'
    )

    if [ -n "$job_vars" ] || [ -n "$service_vars" ]; then
        declare -A vars=()

        SVCMANIFEST=$(mktemp /tmp/gitlab-runner-svc-manifest.XXXXXXXXXX)
        TMPFILES+=("$SVCMANIFEST")
        chmod 600 "$SVCMANIFEST"

        {
            echo "---"
            echo "applications:"
            echo "- name: $container_id"
            echo "  env:"
        } >>"$SVCMANIFEST"
    fi

    if [ -n "$job_vars" ]; then
        while read -r var; do
            read -r key val <<<"$var"
            vars[$key]="$val"
        done <<<"$job_vars"
    fi

    if [ -n "$service_vars" ]; then
        while read -r var; do
            read -r key val <<<"$var"
            vars[$key]="$val"
        done <<<"$service_vars"
    fi

    if [ "${#vars[@]}" -gt 0 ]; then
        for key in "${!vars[@]}"; do
            echo "    $key: ${vars[$key]}" >>"$SVCMANIFEST"
        done

        push_args+=('-f' "$SVCMANIFEST")
        if [ -z "${RUNNER_DEBUG-}" ] || [ "$RUNNER_DEBUG" != "true" ]; then
            push_args+=('--redact-env')
        fi
    fi

    get_start_command push_args "${service_entrypoint[@]}" "${service_command[@]}"

    local docker_user docker_pass
    read -r docker_user docker_pass <<<"$(get_registry_credentials "$image_name")"

    if [ -n "$docker_user" ] && [ -n "$docker_pass" ]; then
        push_args+=('--docker-username' "${docker_user}")
        local -x CF_DOCKER_PASSWORD="${docker_pass}"
    fi

    # TODO - Figure out how to handle non-global memory definition
    cf push "${push_args[@]}"

    # Map route and export a FQDN. We assume apps.internal as the domain.
    cf map-route "$container_id" apps.internal --hostname "$container_id"

    # For use in inter-container communication
    # TODO: propose a devtools/cloudgov 'namespace'
    # TODO: propose rename to, e.g., CG_APP_HOST_X and CG_APP_ID_X
    export "CI_SERVICE_${alias_name}"="${container_id}.apps.internal"
    export "CI_SERVICE_ID_${alias_name}"="${container_id}"
}

start_services() {
    container_id_base="$1"
    ci_job_services="$2"

    if [ -z "$ci_job_services" ]; then
        echo "[cf-driver] No services defined in ci_job_services - Skipping service startup"
        return
    fi

    # GitLab Runner creates JOB_RESPONSE_FILE to provide full job context
    # See: https://docs.gitlab.com/runner/executors/custom.html#job-response
    services=$(jq -rc '.services[]' "$JOB_RESPONSE_FILE")
    job_vars=$(jq -r \
        '.variables[]? | select(.file == false) | select(.key | test("^(?!(CI|GITLAB)_)")) | [.key, .value] | @sh' \
        "$JOB_RESPONSE_FILE")

    for l in $services; do
        # Using jq -er to fail of alias or name are not found
        alias_name=$(echo "$l" | jq -er '.alias | select(.)')
        image_name=$(echo "$l" | jq -er '.name | select(.)')
        container_id="${container_id_base}-svc-${alias_name}"

        # Using jq -r to allow entrypoint, command, and variables to be empty
        service_entrypoint=$(echo "$l" | jq -r '.entrypoint | select(.)[]')
        service_command=$(echo "$l" | jq -r '.command | select(.)[]')

        # start_service will further process the variables, so just compact it
        service_vars=$(echo "$l" | jq -r '.variables[]? | [.key, .value] | @sh')

        start_service "$alias_name" "$container_id" "$image_name" \
            "$service_entrypoint" "$service_command" "$service_vars" "$job_vars"
    done
}

allow_access_to_service() {
    source_app="$1"
    destination_service_app="$2"

    # TODO NOTE: This is foolish and allows all TCP ports for now.
    # This is limiting and sloppy.
    protocol="tcp"
    ports="20-10000"

    cf add-network-policy "$source_app" "$destination_service_app" \
        --protocol "$protocol" --port "$ports"
}

allow_access_to_services() {
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

install_dependencies() {
    container_id="$1"
    local dir="$currentDir/worker-setup"

    echo "[cf-driver] Copying bundle"
    cf_scpr "$container_id" "$dir/bundle"

    echo "[cf-driver] Installing bundle"
    cf_ssh "$container_id" "./bundle/glrw-setup.sh"
}

echo "[cf-driver] re-auth to cloud.gov"
cf orgs &>/dev/null || (cf auth && cf target -o "$WORKER_ORG" -s "$WORKER_SPACE")

if [ "$RUNNER_DEBUG" == "true" ]; then
    echo "[cf-driver] JOB_RESPONSE_FILE ======================================="
    cat "$JOB_RESPONSE_FILE"
    echo "[cf-driver] ======================================= JOB_RESPONSE_FILE"
fi

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
