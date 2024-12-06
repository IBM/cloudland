#!/bin/bash

source tokenrc

secgroups=$(curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/security_groups")
length=$(jq '.security_groups | length' <<<$secgroups)
i=0
while [ $i -lt $length ]; do
	secgroup_id=$(jq -r .security_groups[$i].id <<<$secgroups)
	curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/security_groups/$secgroup_id/rules" | jq .
	let i=$i+1
done
