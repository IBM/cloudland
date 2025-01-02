#!/bin/bash

source tokenrc

instances=$(curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/instances")
length=$(jq '.instances | length' <<<$instances)
i=0
while [ $i -lt $length ]; do
	instance_id=$(jq -r .instances[$i].id <<<$instances)
	curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/instances/$instance_id/interfaces" | jq .
	let i=$i+1
done
