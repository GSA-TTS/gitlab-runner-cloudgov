#!/usr/bin/env bash

usage="
$0: Run SSH through egress-proxy with corkscrew. Must be provided password with either -e, -p, or STDIN.

Usage:
  $0 -h
  $0 -g <app guid> -H <proxy host> -P <proxy port> -F <path to auth file> [-p <ssh password>|-e] <CMD>

Options:
-h 	show help and exit
-g 	GUID of the app to connect with
-H 	egress proxy host
-P 	egress proxy port
-F 	egress proxy basic auth credential file
[-p] 	ssh password
[-e] 	take password from env var 'SSHPASS'
<CMD> 	command to pass to ssh as-is
"

function exitWithUsage() {
  echo "$usage"
  exit "$1"
}

guid=""
proxyHost=""
proxyPort=""
proxyFile=""
sshPass=""

set -e
while getopts ":hg:H:P:F:p:e" opt; do
  case "${opt}" in
  g)
    guid=${OPTARG}
    ;;
  H)
    proxyHost=${OPTARG}
    ;;
  P)
    proxyPort=${OPTARG}
    ;;
  F)
    proxyFile=${OPTARG}
    ;;
  p)
    sshPass=${OPTARG}
    ;;
  e)
    sshPass=${SSHPASS:?"-e used but SSHPASS undefined, run with -h for usage"}
    ;;
  h | *)
    exitWithUsage
    ;;
  esac
done
shift $((OPTIND - 1))

# read sshPass from stdin
if [[ -z "$sshPass" ]]; then
  sshPass=$(cat)
fi

# still no sshPass, exiting
if [[ -z "$sshPass" ]]; then
  echo "error: ssh password is required but none passed with -p, -e or STDIN"
  exitWithUsage 1
fi

if [[ -z "$guid" ]]; then
  echo "error: -g <app guid> is required"
  exitWithUsage 1
fi

if [[ -z "$proxyHost" ]]; then
  echo "error: -H <proxy host> is required"
  exitWithUsage 1
fi

if [[ -z "$proxyPort" ]]; then
  echo "error: -P <proxy port> is required"
  exitWithUsage 1
fi

if [[ -z "$proxyFile" ]]; then
  echo "error: -F <path to auth file> is required"
  exitWithUsage 1
fi

SSHPASS=$sshPass sshpass -e ssh -p 2222 -T \
  -o "StrictHostKeyChecking=no" \
  -o "ProxyCommand corkscrew $proxyHost $proxyPort %h %p $proxyFile" \
  cf:"$guid"/0@ssh.fr.cloud.gov "$@"
