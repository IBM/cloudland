#!/bin/bash

cd `dirname $0`
source ../../cloudrc

[ $# -lt 3 ] && die "$0 <ID> <prefix> <url>"

ID=$1
prefix=$2
url=$3

image_name=image-$ID-$prefix
state=error
mkdir -p $image_cache
image=$image_cache/$image_name
curl -s -k $url -o $image
if [ ! -s "$image" ]; then
    echo "|:-COMMAND-:| $(basename $0) '$ID' '$state' '$format'"
    exit -1
fi

format=$(qemu-img info $image | grep 'file format' | cut -d' ' -f3)
[ "$format" = "qcow2" -o "$format" = "raw" ] && state=downloaded
image_size=$(qemu-img info ${image} | grep 'virtual size:' | cut -d' ' -f5 | tr -d '(')

if [ -z "$wds_address" ]; then
    mv $image ${image}.$format
    state=available
    #sync_target /opt/cloudland/cache/image
else
    get_wds_token
    qemu-img convert -f $format -O raw ${image} ${image}.raw
    format=raw
    uss_id=$(get_uss_gateway)
    uss_service=$(systemctl -a | grep uss | awk '{print $1}')
    cat /etc/systemd/system/$uss_service | grep cloudland
    if [ $? -ne 0 ]; then
        wds_curl PUT "api/v2/sync/wds/uss/$uss_id" '{"action":"add","mount_path":"/opt/cloudland/cache/image"}'
	systemctl restart $uss_service
    fi
    task_id=$(wds_curl "PUT" "api/v2/sync/block/volumes/import" "{\"volname\": \"$image_name\", \"path\": \"${image}.raw\", \"ussid\": \"$uss_id\", \"start_blockid\": 0, \"volsize\": $image_size, \"poolid\": \"$wds_pool_id\", \"num_block\": 0, \"speed\": 8}" | jq -r .task_id)
    state=uploading
    for i in {1..100}; do
        st=$(wds_curl GET "api/v2/sync/block/volumes/tasks/$task_id" | jq -r .task.state)
	[ "$st" = "TASK_COMPLETE" ] && state=uploaded && break
	[ "$st" = "TASK_FAILED" ] && state=failed && break
	sleep 5
    done
    rm -f ${image}
    volume_id=$(wds_curl GET "api/v2/sync/block/volumes?name=$image_name" | jq -r '.volumes[0].id')
    [ -n "$volume_id" ] && state=available
fi
echo "|:-COMMAND-:| $(basename $0) '$ID' '$state' '$format' '$image_size'"
