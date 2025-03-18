#!/usr/bin/env bash
# a little helper script to remind you to set up authentication and pass in the proper cf_user variable

if [ -z "$CF_USER" ] || [ -z "$CF_PASSWORD" ]; then
  echo "ERROR: Export CF_USER and CF_PASSWORD for a user with OrgManager permissions"
  exit 1
fi

TF_VAR_cf_org_manager="$CF_USER" TF_VAR_cf_community_user="$CF_USER" terraform test
