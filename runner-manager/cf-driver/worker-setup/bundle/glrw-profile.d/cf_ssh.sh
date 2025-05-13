#!/bin/sh

cf_ssh() {
    container_id="$1"
    command="$2"
    app_guid=$(cf app "$container_id" --guid)
    SSHPASS=$(cf ssh-code) sshpass -e ssh -p 2222 -T "cf:$app_guid/0@ssh.fr.cloud.gov" "$command"
}
