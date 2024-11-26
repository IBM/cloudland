#!/bin/bash

source tokenrc

instances=$(curl -k -XGET -H "Authorization: bearer $token" 'https://127.0.0.1:8255/api/v1/instances' | jq .)
length=$(jq '.instances | length' <<<$instances)
i=0
while [ $i -lt $length ]; do
	instance_id=$(jq -r .instances[$i].id <<<$instances)
	curl -k -XDELETE -H "Authorization: bearer $token" "https://127.0.0.1:8255/api/v1/instances/$instance_id" | jq .
	let i=$i+1
done
