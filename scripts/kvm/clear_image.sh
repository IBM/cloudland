#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 2 ] && die "$0 <ID> <format>"

ID=$1
format=$2

if [ -z "$wds_address" ]; then
    image=$image_cache/image-${ID}.${format}
    rm -f $image
else
    snapshot=$(wds_curl GET "api/v2/sync/block/snaps" | jq --arg snap image-$ID -r '.snaps | .[] | select(.name == $snap)')
    snapshot_id=$(jq -r .id <<<$snapshot)
    volume_id=$(jq -r .volume_id <<<$snapshot)
    [ -n "$snapshot_id" ] && wds_curl DELETE "api/v2/sync/block/snaps/$snapshot_id?force=false"
    [ -n "$volume_id" ] && wds_curl DELETE "api/v2/sync/block/volumes/$volume_id?force=false"
fi
