#!/bin/bash

source tokenrc

images=$(curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/images" | jq .)
length=$(jq '.images | length' <<<$images)
i=0
while [ $i -lt $length ]; do
	image_id=$(jq -r .images[$i].id <<<$images)
	curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/images/$image_id" | jq .
	let i=$i+1
done
