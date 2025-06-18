#!/usr/bin/env bash

usage() {
    msg="$1"
    status=0
    if [[ -n "$msg" ]]; then
        printf "ERROR: %s\n\n" "$msg" >&2
        status=1
    fi

    cat >&2 <<-EOM
	Usage: $0 [-ps] BASENAME

	Creates a service account with key and a sample application, outputs to testdata dirs.

	Options:
	  -p	Use api.fr.cloud.gov (defaults to fr-stage)
	  -s	Skip creation, only output creds for BASENAME
	EOM

    exit $status
}

dir=$(dirname "$0")
cd "$dir"

cf_api="api.fr-stage.cloud.gov"
app_name="cfd_integration_test_AppGet"
skip_create=

while getopts ":psh" opt; do
    case $opt in
    p)
        cf_api="api.fr.cloud.gov"
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

# create the space deployer, then a key for the deployer
if [[ -z "$skip_create" ]]; then
    cf create-service cloud-gov-service-account space-deployer "$name"-deployer
    cf create-service-key "$name"-deployer "$name"-key
fi

# create a teeny app we can use to test client.AppGet
cf push --no-start -k 8M -m 128M -o busybox -u process -c /bin/sh "$app_name"

out_arr=(
    # get the credentials from key and output
    "$(cf service-key "$name"-deployer "$name"-key | tail +2 |
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
