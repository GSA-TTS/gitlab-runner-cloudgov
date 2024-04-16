#!/usr/bin/env bash
#
# This is sourced by prepare, run, and cleanup

# This name will be long. Hopefully not too long!
CONTAINER_ID="runner-$CUSTOM_ENV_CI_RUNNER_ID-project-$CUSTOM_ENV_CI_PROJECT_ID-concurrent-$CUSTOM_ENV_CI_CONCURRENT_PROJECT_ID-$CUSTOM_ENV_CI_JOB_ID"

# Set a fallback if not set but complain
if [ -v $DEFAULT_JOB_IMAGE ]; then
	DEFAULT_JOB_IMAGE="ubuntu:latest"
	echo "WARNING: DEFAULT_JOB_IMAGE not set! Falling back to ${DEFAULT_JOB_IMAGE}"
fi

# Use a custom image if provided, else fallback to configured default
CUSTOM_ENV_CI_JOB_IMAGE="${CUSTOM_ENV_CI_JOB_IMAGE:=$DEFAULT_JOB_IMAGE}"
