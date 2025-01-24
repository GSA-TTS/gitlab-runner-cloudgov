#!/usr/bin/env bash

org="gsa-tts-devtools-prototyping"
cmd="plan"
sa_username="$(whoami)-glr-sandbox"

usage="
$0: Run terraform commands against a given environment

Usage:
  $0 -h
  $0 [-o <ORG>] [-m <MGMT_SPACE_NAME>] [-u <SERVICE_ACCOUNT_USERNAME>] [-f] [-c <TERRAFORM-CMD>] [-- <EXTRA CMD ARGUMENTS>]

Options:
-h: show help and exit
-o ORG: cloud.gov org. Defaults to $org
-s MGMT_SPACE_NAME: MGMT_SPACE_NAME - required unless 'secrets.auto.tfvars' already exists
-u SERVICE_ACCOUNT_USERNAME: Deployer account name. Defaults to $sa_username
-f: Force, pass -auto-approve to all invocations of terraform
-c TERRAFORM-CMD: command to run. Defaults to $cmd
[<EXTRA CMD ARGUMENTS>]: arguments to pass as-is to terraform
"


mgmt_space=""
force=""
args_to_shift=0

set -e
while getopts ":ho:s:u:fc:" opt; do
  case "$opt" in
    o)
      org=${OPTARG}
      args_to_shift=$((args_to_shift + 2))
      ;;
    s)
      mgmt_space=${OPTARG}
      args_to_shift=$((args_to_shift + 2))
      ;;
    u)
      sa_username=${OPTARG}
      args_to_shift=$((args_to_shift + 2))
      ;;
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

if [[ ! -f secrets.auto.tfvars ]]; then
  if [[ -z "$mgmt_space" ]]; then
    echo "You must specify the management space to use to create the service account"
    echo "$usage"
    exit 1
  fi
  cf target -o "$org" -s "$mgmt_space" || cf login -a api.fr.cloud.gov --sso -o "$org" -s "$mgmt_space"
  cf create-service cloud-gov-service-account space-deployer "$sa_username"
  cf create-service-key "$sa_username" service-account-key
  creds=`cf service-key "$sa_username" service-account-key | tail -n +2 | jq '.credentials'`
  username=`echo "$creds" | jq -r '.username'`
  password=`echo "$creds" | jq -r '.password'`

  cf set-org-role "$username" "$org" OrgManager
  cat > secrets.auto.tfvars << EOF
# generated with terraform.sh, will be cleaned up by terraform.sh when destroying the entire sandbox
#mgmt_space $mgmt_space
#sa_username $sa_username
cf_user     = "$username"
cf_password = "$password"
EOF
else
  # ensure we're logged in via cli
  mgmt_space="$(grep mgmt_space secrets.auto.tfvars | cut -d' ' -f2-)"
  cf spaces &> /dev/null || cf login -a api.fr.cloud.gov --sso -o "$org" -s "$mgmt_space"
fi

terraform init
echo "=============================================================================================================="
echo "= Calling $cmd $force $@ on the sandbox infrastructure"
echo "=============================================================================================================="
terraform "$cmd" $force -compact-warnings "$@"

if [[ "$cmd" = "destroy" ]]; then
  cf target -o "$org" -s "$mgmt_space"
  cf delete-service -f "$(grep sa_username secrets.auto.tfvars | cut -d' ' -f2-)"
  rm secrets.auto.tfvars
fi
