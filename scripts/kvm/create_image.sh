#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 2 ] && die "$0 <ID> <url>"

ID=$1
url=$2

state=error
mkdir -p $image_cache
image=$image_cache/image-$1
curl -s $url -o $image
if [ ! -s "$image" ]; then
    echo "|:-COMMAND-:| $(basename $0) '$ID' '$state' '$format'"
    exit -1
fi

format=$(qemu-img info $image | grep 'file format' | cut -d' ' -f3)
[ "$format" = "qcow2" -o "$format" = "raw" ] && state=downloaded

if [ -z "$wds_address" ]; then
    mv $image ${image}.$format
    #sync_target /opt/cloudland/cache/image
else
    qemu-img convert -f $format -O raw ${image} ${image}.raw
    format=raw
    image_size=$(qemu-img info ${image}.raw | grep 'virtual size:' | cut -d' ' -f5 | tr -d '(')
    uss_id=$(wds_curl GET "api/v2/wds/uss" | jq --arg hname $(hostname -s) -r '.uss_gateways | .[] | select(.server_name == $hname) | .id')
    task_id=$(wds_curl "PUT" "api/v2/sync/block/volumes/import" "{\"volname\": \"image-$ID\", \"path\": \"${image}.raw\", \"ussid\": \"$uss_id\", \"start_blockid\": 0, \"volsize\": $image_size, \"poolid\": \"$wds_pool_id\", \"num_block\": 0, \"speed\": 8}" | jq -r .task_id)
    state=uploading
    for i in {1..100}; do
        st=$(wds_curl GET "api/v2/sync/block/volumes/tasks/$task_id" | jq -r .task.state)
	[ "$st" = "TASK_COMPLETE" ] && state=uploaded && break
	sleep 5
    done
    rm -f ${image}.raw
    volume_id=$(wds_curl GET "api/v2/sync/block/volumes" | jq --arg image image-$ID -r '.volumes | .[] | select(.name == $image) | .id')
    [ -n "$volume_id" ] && state=available
fi
echo "|:-COMMAND-:| $(basename $0) '$ID' '$state' '$format'"
