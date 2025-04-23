#!/usr/bin/env bash
# a little helper script to help pass in the proper cf_org_manager variable

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 your.email@gsa.gov"
  echo "You must be an OrgManager within cf_org_name to call this script"
  exit 1
fi

TF_VAR_cf_org_manager="$1" terraform test
