#!/bin/bash

cd $(dirname $0)
source ../cloudrc

# volume.ID, volume.Size, volume.UUID, iopsLimit, iopsBurst, bpsLimit, bpsBurst, poolID
[ $# -lt 8 ] && echo "$0 <vol_ID> <size> <vol_UUID> <iops_limit> <iops_burst> <bps_limit> <bps_brust> <pool_ID>" && exit -1

vol_ID=$1
size=$2
vol_UUID=$3
iops_limit=$4
iops_burst=$5
bps_limit=$6
bps_burst=$7
pool_ID=$8

state='error'

if [ -z "$wds_address" ]; then
    echo "|:-COMMAND-:| $(basename $0) '$vol_ID' '$state' 'wds_address is not set'"
    exit -1
fi

if [ -z "$pool_ID" ]; then
    pool_ID=$wds_pool_id
fi
if [ -z "$pool_ID" ]; then
    echo "|:-COMMAND-:| $(basename $0) '$vol_ID' '$state' 'pool_ID is not set'"
    exit -1
fi

get_wds_token
state="creating"
let size=$size*1024*1024*1024 # GB to Bytes
# fix wds said: "The volume name cannot start with a number"
vol_WDS_NAME="vol-$vol_ID-$vol_UUID"
result=$(wds_curl "POST" "/api/v2/sync/block/volumes" "{\"phy_pool_id\": \"$pool_ID\", \"name\": \"$vol_WDS_NAME\", \"volume_size\": $size, \"qos\":{\"iops_limit\": $iops_limit, \"iops_burst\": $iops_burst, \"bps_limit\": $bps_limit, \"bps_burst\": $bps_burst}}")
ret_code=$(echo $result | jq -r .ret_code)
if [ "$ret_code" != "0" ]; then
    echo "|:-COMMAND-:| $(basename $0) '$vol_ID' 'error' 'failed to create volume'"
    exit -1
fi
wds_volume_id=$(echo $result | jq -r .id)
if [ -z "$wds_volume_id" ]; then
    echo "|:-COMMAND-:| $(basename $0) '$vol_ID' 'error' 'failed to get volume ID'"
    exit -1
fi
state='available'

echo "|:-COMMAND-:| $(basename $0) '$vol_ID' '$state' 'wds_vhost://$pool_ID/$wds_volume_id'"
