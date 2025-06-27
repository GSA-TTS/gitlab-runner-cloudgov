#!/usr/bin/env bash

cf_api="api.fr-stage.cloud.gov"
cf_api_prod="api.fr.cloud.gov"

basename="wsr-integration"
app_name="cfd_integration_test_AppGet"

usage() {
    msg="$1"
    status=0
    if [[ -n "$msg" ]]; then
        printf "ERROR: %s\n\n" "$msg" >&2
        status=1
    fi

    cat >&2 <<-EOM
	Usage: $0 [-bpf]

	Deletes testing service account, key and sample application.

	Options:
	  -b	Basename to use for service account & key (defaults to $basename)
	  -p	Use $cf_api_prod (defaults to $cf_api)
	  -f	Force deletion without confirmation
	EOM

    exit $status
}

declare -a args

while getopts ":b:pfh" opt; do
    case $opt in
    b)
        basename="$OPTARG"
        ;;
    p)
        cf_api="$cf_api_prod"
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

set -euo pipefail

# check login
cf spaces &>/dev/null || cf login -a "$cf_api" --sso

org=$(cf t | grep org | awk '{print $2}')
if [[ $org != 'sandbox-gsa' ]]; then
    echo "ERROR: you should really probably use this in your sandbox for now"
    exit 1
fi

# delete the teeny app
cf delete -r "${args[@]}" "$app_name"

# delete the deployer and key
cf delete-service-key "${args[@]}" "$basename"-deployer "$basename"-key
cf delete-service "${args[@]}" "$basename"-deployer

echo "WARNING: didn't delete any testdata files, you must remove them manually."
