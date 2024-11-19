#!/bin/bash

# Exit if any command fails
set -e

# The apt buildpack is first, so it installs in the deps/0 directory
PATH="${PATH}:${HOME}/deps/0/bin"

function exit_with_failure() {
    echo "FAILURE: $1"
    exit 1
}

function command_exists() {
    command -v "$1" >/dev/null 2>&1
}

function get_cf_configuration() {
    # Authenticate with Cloud Foundry to allow management of executor app instances
    CF_USERNAME=$(echo "$VCAP_SERVICES" | jq --raw-output --arg tag_name "gitlab-service-account" ".[][] | select(.tags[] == \$tag_name) | .credentials.username")
    CF_PASSWORD=$(echo "$VCAP_SERVICES" | jq --raw-output --arg tag_name "gitlab-service-account" ".[][] | select(.tags[] == \$tag_name) | .credentials.password")
    CF_API=$(echo "$VCAP_APPLICATION" | jq --raw-output ".cf_api")

    if [ -z "$WORKER_ORG" ]; then
        # Use the current CloudFoundry org for workers
        WORKER_ORG=$(echo "$VCAP_APPLICATION" | jq --raw-output ".organization_name")
    fi

    CURRENT_SPACE=$(echo "$VCAP_APPLICATION" | jq --raw-output ".space_name")
    if [ -z "$WORKER_SPACE" ]; then
        WORKER_SPACE="$CURRENT_SPACE"
    fi

    if [ "$WORKER_SPACE" = "$CURRENT_SPACE" ]; then
        echo "WARNING: Use the same space for the runner manager and workers is not recommended: Configure WORKER_SPACE (worker-space) to use a different space"
    fi
}

function auth_cf() {
    cf api "$CF_API"
    cf auth "$CF_USERNAME" "$CF_PASSWORD"
    cf target -o "$WORKER_ORG" -s "$WORKER_SPACE"
}

function get_object_store_configuration() {
    # OBJECT_STORE_INSTANCE holds the service name for our brokered S3 bucket.
    # This pulls the config out into GitLab Runner environment variables.
    export CACHE_TYPE="s3"
    export CACHE_SHARED="true"
    export CACHE_PATH="$RUNNER_NAME" # Allows for multiple runner sets to use one bucket - ONLY USE FOR RUNNERS AT THE SAME SECURITY LEVEL

    export CACHE_S3_SERVER_ADDRESS=$(echo "$VCAP_SERVICES" | jq --raw-output --arg service_name "$OBJECT_STORE_INSTANCE" ".[][] | select(.name == \$service_name) | .credentials.fips_endpoint")
    export CACHE_S3_BUCKET_LOCATION=$(echo "$VCAP_SERVICES" | jq --raw-output --arg service_name "$OBJECT_STORE_INSTANCE" ".[][] | select(.name == \$service_name) | .credentials.region")
    export CACHE_S3_BUCKET_NAME=$(echo "$VCAP_SERVICES" | jq --raw-output --arg service_name "$OBJECT_STORE_INSTANCE" ".[][] | select(.name == \$service_name) | .credentials.bucket")
    export CACHE_S3_ACCESS_KEY=$(echo "$VCAP_SERVICES" | jq --raw-output --arg service_name "$OBJECT_STORE_INSTANCE" ".[][] | select(.name == \$service_name) | .credentials.access_key_id")
    export CACHE_S3_SECRET_KEY=$(echo "$VCAP_SERVICES" | jq --raw-output --arg service_name "$OBJECT_STORE_INSTANCE" ".[][] | select(.name == \$service_name) | .credentials.secret_access_key")
}

echo "Getting CloudFoundry configuration"
get_cf_configuration

echo "Setting up CloudFoundry API access"
auth_cf

if [ -n "${OBJECT_STORE_INSTANCE:+x}" ]; then
    echo "Getting object store configuration for cache and artifacts"
    get_object_store_configuration

    echo "Cache configured to use AWS S3 \"$CACHE_S3_BUCKET_NAME/$CACHE_PATH\""
fi

# Allow safe re-running... Idempotent-ial.
if pgrep 'gitlab-runner' > /dev/null ; then
    echo "DevTools Runner - Manager appears to already be running"
else
    echo "Registering Devtools Runner - Manager with name $RUNNER_NAME"
    if gitlab-runner register; then
        echo "DevTools Runner - Manager successfully registered"
    else
        exit_with_failure "DevTools Runner $RUNNER_NAME not registered"
    fi
fi
