#!/usr/bin/env bash

cmd="plan"

usage="
$0: Run terraform commands against a given environment

Usage:
  $0 -h
  $0 [-f] [-c <TERRAFORM-CMD>] [-r] [-- <EXTRA CMD ARGUMENTS>]

Options:
-h: show help and exit
-f: Force, pass -auto-approve to all invocations of terraform
-r: Replace worker space, useful to reset after using PRESERVE_WORKER
-c TERRAFORM-CMD: command to run. Defaults to $cmd
[<EXTRA CMD ARGUMENTS>]: arguments to pass as-is to terraform
"


force=""
args_to_shift=0
declare -a optional_args

set -e
while getopts ":hfrc:" opt; do
  case "$opt" in
    f)
      force="-auto-approve"
      args_to_shift=$((args_to_shift + 1))
      ;;
    c)
      cmd=${OPTARG}
      args_to_shift=$((args_to_shift + 2))
      ;;
    r)
      optional_args+=(-replace=module.sandbox-runner.module.worker_space.cloudfoundry_space.space)
      optional_args+=(-replace="module.sandbox-runner.module.worker_space.cloudfoundry_security_group_space_bindings.security_group_bindings[\"trusted_local_networks_egress\"]")
      args_to_shift=$((args_to_shift + 1))
      ;;
    h)
      echo "$usage"
      exit 0
      ;;
  esac
done

shift $args_to_shift
if [[ "$1" = "--" ]]; then
  shift 1
fi

dev_org="cloud-gov-devtools-development"
cf org $dev_org &> /dev/null || cf login -a api.fr-stage.cloud.gov --sso -o $dev_org

terraform init
echo "=============================================================================================================="
echo "= Calling $cmd $force " "${optional_args[@]}" " " "$@" " on the sandbox infrastructure"
echo "=============================================================================================================="
terraform "$cmd" $force -compact-warnings "${optional_args[@]}" "$@"
