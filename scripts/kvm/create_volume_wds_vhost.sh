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

task_id=$(wds_curl "POST" "/api/v2/block/volumes" "{\"phy_pool_id\": \"$pool_ID\", \"name\": \"$vol_UUID\", \"volume_size\": $size, \"qos\":{\"iops_limit\": $iops_limit, \"iops_burst\": $iops_burst, \"bps_limit\": $bps_limit, \"bps_burst\": $bps_burst}}" | jq -r .task_id)
state="creating"

for i in {1..100}; do
    st=$(wds_curl GET "api/v2/block/volumes/tasks/$task_id" | jq -r .task.state)
    [ "$st" = "TASK_COMPLETE" ] && state="created" && break
    sleep 5
done

wds_volume_id=$(wds_curl GET "api/v2/sync/block/volumes?name=$vol_UUID" | jq --arg volname $vol_UUID -r '.volumes | .[] | select(.name == $volname) | .id')

[ $? -eq 0 ] && state='available'

echo "|:-COMMAND-:| $(basename $0) '$vol_ID' '$state' 'wds_vhost://$pool_ID/$wds_volume_id'"
