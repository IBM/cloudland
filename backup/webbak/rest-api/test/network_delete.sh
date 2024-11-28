#!/usr/bin/env bash
network_id=$1

source base.sh
source token.sh
cmd="curl -X DELETE ${host}${network_endpoint}/$1 -s  -H 'X-Auth-Token: ${token}'"
echo $cmd
result=$(eval $cmd)
echo $result
