#!/usr/bin/env bash

cmd="plan"

usage="
$0: Run terraform commands against a given environment

Usage:
  $0 -h
  $0 [-f] [-c <TERRAFORM-CMD>] [-- <EXTRA CMD ARGUMENTS>]

Options:
-h: show help and exit
-f: Force, pass -auto-approve to all invocations of terraform
-c TERRAFORM-CMD: command to run. Defaults to $cmd
[<EXTRA CMD ARGUMENTS>]: arguments to pass as-is to terraform
"


force=""
args_to_shift=0

set -e
while getopts ":hfc:" opt; do
  case "$opt" in
    f)
      force="-auto-approve"
      args_to_shift=$((args_to_shift + 1))
      ;;
    c)
      cmd=${OPTARG}
      args_to_shift=$((args_to_shift + 2))
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

cf org cloud-gov-devtools-development &> /dev/null || cf login -a api.fr-stage.cloud.gov --sso

terraform init
echo "=============================================================================================================="
echo "= Calling $cmd $force $@ on the sandbox infrastructure"
echo "=============================================================================================================="
terraform "$cmd" $force -compact-warnings "$@"
