#!/bin/sh
#
# This runs before gitlab-runner, allowing for additional pre-configuration.
#
# If this is Ash, let's pretend we are bash...
ENABLE_ASH_BASH_COMPAT=1

set -e

if [ "$OBJECT_STORE_INSTANCE" -ne "" ]; then
    # OBJECT_STORE_INSTANCE holds the service name for our brokered S3 bucket.
    # This pulls the config out into GitLab Runner environment variables.
    SCONFIG="$(echo "$VCAP_SERVICES" | jq --raw-output --arg service_name "$OBJECT_STORE_INSTANCE" ".[][] | select(.name == \$service_name)")"

    export CACHE_TYPE="s3"
    export CACHE_SHARED="true"
    export CACHE_PATH="$RUNNER_NAME" # Allow for multiple runners to use the bucket - ONLY USE FOR RUNNERS AT THE SAME SECURITY LEVEL

    export CACHE_S3_SERVER_ADDRESS="$(echo "$SCONFIG" | jq --raw-output ".credentials.fips_endpoint")"
    export CACHE_S3_BUCKET_LOCATION="$(echo "$SCONFIG" | jq --raw-output ".credentials.region")"
    export CACHE_S3_BUCKET_NAME="$(echo "$SCONFIG" | jq --raw-output ".credentials.bucket")"
    export CACHE_S3_ACCESS_KEY="$(echo "$SCONFIG" | jq --raw-output ".credentials.access_key_id")"
    export CACHE_S3_SECRET_KEY="$(echo "$SCONFIG" | jq --raw-output ".credentials.secret_access_key")"

    echo "Cache configured to use AWS S3 \"$CACHE_S3_BUCKET_NAME/$CACHE_PATH\""
fi

