#!/bin/bash

source tokenrc

flavors=$(curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/flavors" | jq .)
length=$(jq '.flavors | length' <<<$flavors)
i=0
while [ $i -lt $length ]; do
	flavor_name=$(jq -r .flavors[$i].name <<<$flavors)
	curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/flavors/$flavor_name" | jq .
	let i=$i+1
done
