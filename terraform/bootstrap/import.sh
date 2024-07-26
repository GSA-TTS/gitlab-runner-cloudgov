#!/usr/bin/env bash

read -p "Are you sure you want to import terraform state (y/n)? " verify

if [[ $verify == "y" ]]; then
  echo "Importing bootstrap state"
  ./run.sh import module.s3.cloudfoundry_service_instance.bucket e6bbad10-b4f9-4c33-86e2-69ae8ce67cc4
  ./run.sh import cloudfoundry_service_key.bucket_creds 
  ./run.sh plan
else
  echo "Not importing bootstrap state"
fi
