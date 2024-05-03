#!/usr/bin/env bash

set -euo pipefail

# trap any error, and mark it as a system failure.
trap 'exit $SYSTEM_FAILURE_EXIT_CODE' ERR

# Prepare a runner executor application in CloudFoundry

currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "${currentDir}/base.sh" # Get variables from base.
if [ -z "$WORKER_MEMORY" ]; then
    WORKER_MEMORY="512M"
fi

start_container () {
    if cf app --guid "$CONTAINER_ID" >/dev/null 2>/dev/null ; then
        echo 'Found old instance of runner executor, deleting'
        cf delete "$CONTAINER_ID"
    fi

    cf push "$CONTAINER_ID" -f "${currentDir}/worker-manifest.yml" \
       --docker-image "$CUSTOM_ENV_CI_JOB_IMAGE" -m "$WORKER_MEMORY" \
       -var "object_store_instance=${OBJECT_STORE_INSTANCE}"
}

install_dependencies () {
    # Build a command to try and install git and git-lfs on common distros.
    # Of course, RedHat/UBI will need more help to add RPM repos with the correct
    # version. TODO - RedHat support
    echo "[cf-driver] Ensuring git, git-lfs, and curl are installed"
    cf ssh "$CONTAINER_ID" -c '(which git && which git-lfs && which curl) || \
                               (which apk && apk add git git-lfs curl) || \
                               (which apt-get && apt-get update && apt-get install -y git git-lfs curl) || \
                               (echo "Required packages missing and I do not know what to do about it" && exit 1)'

    # gitlab-runner-helper includes a limited subset of gitlab-runner functionality
    # plus Git and Git-LFS. https://s3.dualstack.us-east-1.amazonaws.com/gitlab-runner-downloads/latest/index.html
    #
    # Install gitlab-runner-helper binary since we need to manage cache/artifacts.
    # Symlinks gitlab-runner to avoid having to alter more of the executor.
    # TODO: Pin the version and support more arches than X86_64
    echo "[cf-driver] Installing gitlab-runner-helper"
    cf ssh "$CONTAINER_ID" -c 'curl -L --output /usr/bin/gitlab-runner-helper \
                               "https://s3.dualstack.us-east-1.amazonaws.com/gitlab-runner-downloads/latest/binaries/gitlab-runner-helper/gitlab-runner-helper.x86_64"; \
                               chmod +x /usr/bin/gitlab-runner-helper; \
                               ln -s /usr/bin/gitlab-runner-helper /usr/bin/gitlab-runner'
}

echo "[cf-driver] Starting $CONTAINER_ID with image $CUSTOM_ENV_CI_JOB_IMAGE"
start_container

echo "[cf-driver] Installing dependencies into $CONTAINER_ID"
install_dependencies

echo "[cf-driver] $CONTAINER_ID preparation complete"
