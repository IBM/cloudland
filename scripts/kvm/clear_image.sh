#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 3 ] && die "$0 <ID> <prefix> <format>"

ID=$1
prefix=$2
format=$3

image_name=image-${ID}-${prefix}.${format}
if [ -z "$wds_address" ]; then
    image=$image_cache/$iname_name
    rm -f $image
else
    volume_id=$(wds_curl GET "api/v2/sync/block/volumes" | jq --arg name $image_name -r '.volumes | .[] | select(.name == $name) | .id')
    if [ -n "$volume_id" ]; then
        snapshots=$(wds_curl GET "api/v2/sync/block/snaps" | jq --arg volume_id $volume_id -r '.snaps | .[] | select(.volume_id == $volume_id) | .id')
    else
	snapshots=$(wds_curl GET "api/v2/sync/block/snaps" | jq --arg name $image_name -r '.snaps | .[] | select(.name | startswith($name)) | .id')
    fi
    for snapshot in $snapshots; do
        wds_curl DELETE "api/v2/sync/block/snaps/$snapshot?force=false"
    done
    [ -n "$volume_id" ] && wds_curl DELETE "api/v2/sync/block/volumes/$volume_id?force=false"
fi
