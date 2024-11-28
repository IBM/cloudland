#!/bin/bash

source tokenrc

subnets=$(curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/subnets" | jq .)
length=$(jq '.subnets | length' <<<$subnets)
i=0
while [ $i -lt $length ]; do
	subnet_id=$(jq -r .subnets[$i].id <<<$subnets)
	curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/subnets/$subnet_id" | jq .
	let i=$i+1
done
