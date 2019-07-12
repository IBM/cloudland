#!/usr/bin/env bash
name=$1
networkType=$2
vlandID=$3

source base.sh
source token.sh
create_network_body=$(jq  ".network.name = \"$name\" | .network.\"provider:network_type\" = \"$networkType\" | .network.\"provider:segmentation_id\" = \"$vlandID\"" ./test_data/network.json)
cmd="curl  ${host}${network_endpoint}  -s  -H 'X-Auth-Token: ${token}'"
result=$(eval $cmd)
echo $result
