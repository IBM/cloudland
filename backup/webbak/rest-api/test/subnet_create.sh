#!/usr/bin/env bash
name=$1
cidr=$2
network_id=$3
[ $# -lt 3 ] && echo "$0 <name> <cidr-192.168.0.1/24> <network_id>" && exit -1
source base.sh
source token.sh
create_subnet_body=$(jq  ".subnet.name = \"${name}\" | .subnet.cidr = \"${cidr}\" | .subnet.network_id = \"${network_id}\"" ./test_data/subnet.json)
cmd="curl  ${host}${subnet_endpoint} -X POST  -s -v  -d '$create_subnet_body' -H 'X-Auth-Token: ${token}'"
echo $cmd
result=$(eval $cmd)
echo $result
