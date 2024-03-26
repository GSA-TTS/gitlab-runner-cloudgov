#!/bin/bash

# The apt buildpack is first, so it installs in the deps/0 directory
PATH="${PATH}:${HOME}/deps/0/bin"

function exit_with_failure() {
    echo "FAILURE: $1"
    exit 1
}

function command_exists() {
    command -v "$1" >/dev/null 2>&1
}

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
    #   1. Check VCAP_SERVICES for service account credentials
    #   2. Use those credentials to log into cloud.gov
fi

echo "Registering GitLab Runner with name $RUNNER_NAME"
# TODO: Use a template to set up the custom runner info
#   https://docs.gitlab.com/runner/register/index.html#register-with-a-configuration-template
if gitlab-runner register --non-interactive \
    --url "$CI_SERVER_URL" \
    --token "$AUTHENTICATION_TOKEN" \
    --executor "$RUNNER_EXECUTOR" \
    --description "$RUNNER_NAME"; then
    echo "GitLab Runner successfully registered"
else
    exit_with_failure "GitLab Runner not registered"
fi
