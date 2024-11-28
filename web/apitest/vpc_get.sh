#!/bin/bash

source tokenrc

vpcs=$(curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/vpcs" | jq .)
length=$(jq '.vpcs | length' <<<$vpcs)
i=0
while [ $i -lt $length ]; do
	vpc_id=$(jq -r .vpcs[$i].id <<<$vpcs)
	curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/vpcs/$vpc_id" | jq .
	let i=$i+1
done
