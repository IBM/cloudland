#!/usr/bin/env bash

source base.sh
source token.sh
cmd="curl  ${host}${network_endpoint}  -s  -H 'X-Auth-Token: ${token}'"
result=$(eval $cmd)
echo $result
