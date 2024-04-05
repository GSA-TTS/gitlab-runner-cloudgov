!/usr/bin/env bash

# Prepare a runner executor application in CloudFoundry

currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source ${currentDir}/base.sh # Get variables from base.

set -eo pipefail

# trap any error, and mark it as a system failure.
trap "exit $SYSTEM_FAILURE_EXIT_CODE" ERR

start_executor_app () {
    if cf app --guid "$CONTAINER_ID" >/dev/null 2>/dev/null ; then
        echo 'Found old instance of runner executor, deleting'
        cf delete "$CONTAINER_ID"
    fi

    cf push "$CONTAINER_ID" --docker-image "$CUSTOM_ENV_CI_JOB_IMAGE" --no-route --health-check-style process
}

install_dependencies () {
    # Install Git LFS, git comes pre installed with ubuntu image.
    cf ssh "$CONTAINER_ID" -c 'curl -s "https://packagecloud.io/install/repositories/github/git-lfs/script.deb.sh" | sudo bash'
    cf ssh "$CONTAINER_ID" -c "apt-get install -y git-lfs"

    # Install gitlab-runner binary since we need for cache/artifacts.
    cf ssh "$CONTAINER_ID" -c 'curl -L --output /usr/local/bin/gitlab-runner "https://gitlab-runner-downloads.s3.amazonaws.com/latest/binaries/gitlab-runner-linux-amd64"'
    cf ssh "$CONTAINER_ID" -c "chmod +x /usr/local/bin/gitlab-runner"
}

echo "Starting $CONTAINER_ID with image $CCUSTOM_ENV_CI_JOB_IMAGE"
start_container

# XXX -I don't like slapping things onto every image - Let's see what breaks without it
#echo "Installing dependencies into $CONTAINER_ID"
#install_dependencies

echo "$CONTAINER_ID preparation complete"
