#!/bin/bash

source tokenrc

instances=$(curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/instances")
length=$(jq '.instances | length' <<<$instances)
i=0
while [ $i -lt $length ]; do
	instance_id=$(jq -r .instances[$i].id <<<$instances)
	interfaces=$(curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/instances/$instance_id/interfaces")
	len=$(jq '.interfaces | length' <<<$interfaces)
	j=0
	while [ $j -lt $len ]; do
		interface_id=$(jq -r .interfaces[$j].id <<<$interfaces)
		echo interface_id: $interface_id
		curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/instances/$instance_id/interfaces/$interface_id" | jq .
		let j=$j+1
	done
	let i=$i+1
done
