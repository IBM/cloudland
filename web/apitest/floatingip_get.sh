#!/bin/bash

source tokenrc

floating_ips=$(curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/floating_ips" | jq .)
length=$(jq '.floating_ips | length' <<<$floating_ips)
i=0
while [ $i -lt $length ]; do
	floating_ip_id=$(jq -r .floating_ips[$i].id <<<$floating_ips)
	curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/floating_ips/$floating_ip_id" | jq .
	let i=$i+1
done
