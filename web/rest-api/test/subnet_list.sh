#!/usr/bin/env bash
source base.sh
source token.sh
cmd="curl  ${host}${subnet_endpoint} -s -H 'X-Auth-Token: ${token}'"
#echo $cmd
result=$(eval $cmd)
echo $result
