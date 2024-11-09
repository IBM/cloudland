#!/usr/bin/env bash
[ $# -lt 1 ] && echo "$0 <subnet-ID>  " && exit -1
subnet_ID=$1
source base.sh
source token.sh
cmd="curl  ${host}${subnet_endpoint}/${subnet_ID} -X DELETE  -s -v  -H 'X-Auth-Token: ${token}'"
echo $cmd
result=$(eval $cmd)
echo $result
