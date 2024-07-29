#!/usr/bin/env bash
#
# This is sourced by prepare, run, and cleanup

# This name will be long. Hopefully not too long!
# Any changes to this pattern need to be mirrored in .gitlab-ci.yml when
# used to prefix service names.
CONTAINER_ID="glrw-r$CUSTOM_ENV_CI_RUNNER_ID-p$CUSTOM_ENV_CI_PROJECT_ID-c$CUSTOM_ENV_CI_CONCURRENT_PROJECT_ID-j$CUSTOM_ENV_CI_JOB_ID"

# Set a fallback if not set but complain
if [ -z "$DEFAULT_JOB_IMAGE" ]; then
    DEFAULT_JOB_IMAGE="ubuntu:latest"
    echo "WARNING: DEFAULT_JOB_IMAGE not set! Falling back to ${DEFAULT_JOB_IMAGE}"
fi

# Complain if no Docker Hub credentials so we aren't bad neighbors
if [ -z "$DOCKER_HUB_USER" ] || [ -z "$DOCKER_HUB_TOKEN" ]; then
    echo "WARNING: Docker Hub credentials not set! Falling back to public access which could result in rate limiting."
fi

# Use a custom image if provided, else fallback to configured default
CUSTOM_ENV_CI_JOB_IMAGE="${CUSTOM_ENV_CI_JOB_IMAGE:=$DEFAULT_JOB_IMAGE}"
