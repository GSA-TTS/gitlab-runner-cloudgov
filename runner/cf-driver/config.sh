#!/usr/bin/env bash
#
# Generate custom configuration JSON - https://docs.gitlab.com/runner/executors/custom.html#config
# STDOUT needs to be well formed JSON or this gets ignored.

TLD=".apps.internal"

currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "${currentDir}/base.sh" # Get variables from base.

SERVICE_PREFIX="${CONTAINER_ID}-svc-"

CUSTOM_ENVIRONMENT=()

if [ -n "$CUSTOM_ENV_CI_JOB_SERVICES" ]; then
    CUSTOM_ENVIRONMENT+=$(echo "$CUSTOM_ENV_CI_JOB_SERVICES" | jq ".[] | {(\"CI_SERVICE_\" + .alias): (\"${SERVICE_PREFIX}\" + .alias + \"$TLD\")}")
fi

if [[ -z ${CUSTOM_ENVIRONMENT[@]} ]]; then
    exit 0
fi

echo "[cf-driver] Adding configuration from config.sh" 1>&2

cat << EOS
{
  "job_env" : {
    "CUSTOM_ENVIRONMENT": $( echo ${CUSTOM_ENVIRONMENT[@]} | jq -s 'add' )
  }
}
EOS
