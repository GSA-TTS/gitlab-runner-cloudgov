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

# Authenticate with Cloud Foundry
CF_USERNAME=$(echo "$VCAP_SERVICES" | jq --raw-output --arg tag_name "gitlab-service-account" ".[]
[] | select(.tags[] == \$tag_name) | .credentials.username")
CF_PASSWORD=$(echo "$VCAP_SERVICES" | jq --raw-output --arg tag_name "gitlab-service-account" ".[][] | select(.tags[] == \$tag_name) | .credentials.password")
CF_API="https://api.fr.cloud.gov"
cf api "${CF_API}"
cf auth "${CF_USERNAME}" "${CF_PASSWORD}"

# Install the App Autoscaler CLI plugin (can we do this at staging?)
cf install-plugin -r CF-Community "app-autoscaler-plugin"

if [ -z "$CI_SERVER_URL" ]; then
    exit_with_failure 'CI_SERVER_URL is missing'
fi
if [ -z "$RUNNER_NAME" ]; then
    exit_with_failure 'RUNNER_NAME is missing'
fi
if [ -z "$AUTHENTICATION_TOKEN" ]; then
    exit_with_failure 'AUTHENTICATION_TOKEN is missing'
fi
if [ -z "$RUNNER_EXECUTOR" ]; then
    export RUNNER_EXECUTOR="shell"

    # TODO: Default instead to custom runner
fi

echo "Registering GitLab Runner with name $RUNNER_NAME"
# TODO: Use a template to set up the custom runner info
#   https://docs.gitlab.com/runner/register/index.html#register-with-a-configuration-template
if gitlab-runner register --non-interactive \
    --url "$CI_SERVER_URL" \
    --token "$AUTHENTICATION_TOKEN" \
    --executor "$RUNNER_EXECUTOR" \
    --name "$RUNNER_NAME"; then
    echo "GitLab Runner successfully registered"
else
    exit_with_failure "GitLab Runner not registered"
fi
