#!/usr/bin/env bash

usage() {
    msg="$1"
    status=0
    if [[ -n "$msg" ]]; then
        printf "ERROR: %s\n\n" "$msg" >&2
        status=1
    fi

    cat >&2 <<-EOM
	Usage: $0 [-pf] BASENAME

	Deletes testing service account, key and sample application.

	Options:
	  -p	Use api.fr.cloud.gov (defaults to fr-stage)
	  -f	Force deletion without confirmation
	EOM

    exit $status
}

cf_api="api.fr-stage.cloud.gov"
app_name="cfd_integration_test_AppGet"
declare -a args

while getopts ":pfh" opt; do
    case $opt in
    p)
        cf_api="api.fr.cloud.gov"
        ;;
    f)
        args+=("-f")
        ;;
    h)
        usage
        ;;
    \?)
        usage "unknown option '-$OPTARG'"
        ;;
    esac
done
shift $((OPTIND - 1))

name="$1"

set -euo pipefail

# check login
cf spaces &>/dev/null || cf login -a "$cf_api" --sso

org=$(cf t | grep org | awk '{print $2}')
if [[ $org != 'sandbox-gsa' ]]; then
    echo "ERROR: you should really probably use this in your sandbox for now"
    exit 1
fi

# check name defined, prompt if not
if [[ -z "$name" ]]; then
    read -rei "wsr-integration" -p "Please input a basename for the service account: " name
fi

# delete the teeny app
cf delete -r "${args[@]}" "$app_name"

# delete the deployer and key
cf delete-service-key "${args[@]}" "$name"-deployer "$name"-key
cf delete-service "${args[@]}" "$name"-deployer

echo "WARNING: didn't delete any testdata files, you must remove them manually."
