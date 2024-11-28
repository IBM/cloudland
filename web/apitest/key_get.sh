#!/bin/bash

source tokenrc

keys=$(curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/keys" | jq .)
length=$(jq '.keys | length' <<<$keys)
i=0
while [ $i -lt $length ]; do
	key_id=$(jq -r .keys[$i].id <<<$keys)
	curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/keys/$key_id" | jq .
	let i=$i+1
done
