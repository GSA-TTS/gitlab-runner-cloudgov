#!/usr/bin/env bash

cf_api="api.fr-stage.cloud.gov"
cf_api_prod="api.fr.cloud.gov"
skip_create=

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
	Usage: $0 [-bps]

	Creates a service account with key and a sample application, outputs to testdata dirs.

	Options:
	  -b	Basename to use for service account & key (defaults to $basename)
	  -p	Use $cf_api_prod (defaults to $cf_api)
	  -s	Skip creation and only get & output credentials
	EOM

    exit $status
}

dir=$(dirname "$0")
cd "$dir"

while getopts ":b:psh" opt; do
    case $opt in
    b)
        basename="$OPTARG"
        ;;
    p)
        cf_api="$cf_api_prod"
        ;;
    s)
        skip_create="true"
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

# create the space deployer, then a key for the deployer
if [[ -z "$skip_create" ]]; then
    cf create-service cloud-gov-service-account space-deployer "$basename"-deployer
    cf create-service-key "$basename"-deployer "$basename"-key
fi

# create a teeny app we can use to test client.AppGet
cf push -k 8M -m 128M -o busybox -u process -c /bin/sh "$app_name"

out_arr=(
    # get the credentials from key and output
    "$(cf service-key "$basename"-deployer "$basename"-key | tail +2 |
        jq -r ".credentials | .username,.password")"
    # get target org & space
    "$(cf t | tail -2 | awk '{print $2}')"
    "$app_name"
)

out_str=$(
    IFS=$'\n'
    echo "${out_arr[*]}"
)

echo "$out_str" >../cloudgov/testdata/.cloudgov_creds
echo "$out_str" >../cmd/drive/testdata/.cloudgov_creds
