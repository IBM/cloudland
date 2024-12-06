#!/bin/bash

source tokenrc

secgroups=$(curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/security_groups")
length=$(jq '.security_groups | length' <<<$secgroups)
i=0
while [ $i -lt $length ]; do
	secgroup_id=$(jq -r .security_groups[$i].id <<<$secgroups)
	secrules=$(curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/security_groups/$secgroup_id/rules")
	len=$(jq '.security_rules | length' <<<$secrules)
	j=0
	while [ $j -lt $len ]; do
		secrule_id=$(jq -r .security_rules[$j].id <<<$secrules)
		echo secrule_id: $secrule_id
		curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/security_groups/$secgroup_id/rules/$secrule_id" | jq .
		let j=$j+1
	done
	let i=$i+1
done
