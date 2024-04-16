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

cf api "$CF_API"
cf auth "$CF_USERNAME" "$CF_PASSWORD"
cf target -o "$WORKER_ORG" -s "$WORKER_SPACE"

echo "Registering GitLab Runner with name $RUNNER_NAME"

# Runner initial configuration is managed by envvars set in manifest.yaml and
# above.
if gitlab-runner register; then
    echo "GitLab Runner successfully registered"
else
    exit_with_failure "GitLab Runner not registered"
fi
