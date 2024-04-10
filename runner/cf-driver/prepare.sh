#!/usr/bin/env bash

# Prepare a runner executor application in CloudFoundry

currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source ${currentDir}/base.sh # Get variables from base.
if [ -z "$WORKER_MEMORY" ]; then
    WORKER_MEMORY="512M"
fi

set -eo pipefail

# trap any error, and mark it as a system failure.
trap "exit $SYSTEM_FAILURE_EXIT_CODE" ERR

start_container () {
    if cf app --guid "$CONTAINER_ID" >/dev/null 2>/dev/null ; then
        echo 'Found old instance of runner executor, deleting'
        cf delete "$CONTAINER_ID"
    fi

    cf push "$CONTAINER_ID" --docker-image "$CUSTOM_ENV_CI_JOB_IMAGE" --no-route --health-check-style process --memory "$WORKER_MEMORY"
}

install_dependencies () {
    # Build a command to try and install git and git-lfs on common distros.
    # Of course, RedHat/UBI will need more help to add RPM repos with the correct
    # version. RedHat support TODO
    cf ssh "$CONTAINER_ID" -c 'which apk && apk add git && apk add git-lfs; which apt-get && apt-get update && apt-get install -y git git-lfs'

    # gitlab-runner-helper includes a limited subset of gitlab-runner functionality
    # plus Git and Git-LFS. https://s3.dualstack.us-east-1.amazonaws.com/gitlab-runner-downloads/latest/index.html
    #
    # Install gitlab-runner-helper binary since we need for cache/artifacts. 
    # TODO: Pin the version and support more arches than X86_64
    cf ssh "$CONTAINER_ID" -c 'curl -L --output /usr/bin/gitlab-runner-helper "https://s3.dualstack.us-east-1.amazonaws.com/gitlab-runner-downloads/latest/binaries/gitlab-runner-helper/gitlab-runner-helper.x86_64"; chmod +x /usr/bin/gitlab-runner-helper'
}

echo "Starting $CONTAINER_ID with image $CUSTOM_ENV_CI_JOB_IMAGE"
start_container

echo "Installing dependencies into $CONTAINER_ID"
install_dependencies

echo "$CONTAINER_ID preparation complete"
