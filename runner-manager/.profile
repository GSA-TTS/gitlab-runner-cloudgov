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

function setup_proxy_access() {
    EGRESS_CREDENTIALS=$(echo "$VCAP_SERVICES" | jq --arg service_name "$PROXY_CREDENTIAL_INSTANCE" ".[][] | select(.name == \$service_name) | .credentials")
    if [ -n "$EGRESS_CREDENTIALS" ]; then
        echo "Configuring HTTPS_PROXY environment variable"
        export https_proxy=$(echo "$EGRESS_CREDENTIALS" | jq --raw-output ".https_uri")
        export http_proxy=$(echo "$EGRESS_CREDENTIALS" | jq --raw-output ".http_uri")
        ssh_proxy_host=$(echo "$EGRESS_CREDENTIALS" | jq --raw-output ".domain")
        ssh_proxy_port=$(echo "$EGRESS_CREDENTIALS" | jq --raw-output ".http_port")
        mkdir -p /home/vcap/.ssh
        cat << EOC > /home/vcap/.ssh/config
Host $CG_SSH_HOST
    StrictHostKeyChecking accept-new
    ProxyCommand corkscrew $ssh_proxy_host $ssh_proxy_port %h %p /home/vcap/.ssh/ssh_proxy.auth
EOC
        echo "$EGRESS_CREDENTIALS" | jq --raw-output ".cred_string" > /home/vcap/.ssh/ssh_proxy.auth
        chmod 0600 /home/vcap/.ssh/ssh_proxy.auth
    else
        echo "WARNING: Could not configure the egress proxy"
    fi
}

function get_cf_configuration() {
    # Authenticate with Cloud Foundry to allow management of executor app instances
    cf api $(echo "$VCAP_APPLICATION" | jq --raw-output ".cf_api")

    if [ -z "$WORKER_ORG" ]; then
        # Use the current CloudFoundry org for workers
        export WORKER_ORG=$(echo "$VCAP_APPLICATION" | jq --raw-output ".organization_name")
    fi
}

function auth_cf() {
    cf auth
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

echo "Setting up https_proxy"
setup_proxy_access

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
    echo "GitLab Runner appears to already be running"
else
    echo "Registering GitLab Runner with name $RUNNER_NAME"
    if gitlab-runner register; then
        sed -e "s/concurrent = 1$/concurrent = $RUNNER_CONCURRENCY/" -i.bak .gitlab-runner/config.toml
        echo "GitLab Runner successfully registered"
    else
        exit_with_failure "GitLab Runner not registered"
    fi
fi
