#!/bin/bash

cd $(dirname $0)
source ../cloudrc

source_hyper=$1
volumes=$(cat)
nvolume=$(jq length <<< $volumes)
src_uss_id=$(get_uss_gateway $source_hyper)
i=0
while [ $i -lt $nvolume ]; do
    read -d'\n' -r vol_ID volume_id device booting < <(jq -r ".[$i].id, .[$i].uuid, .[$i].device, .[$i].booting" <<<$volumes)
    echo vol_ID:$vol_ID volume_id:$volume_id device:$device booting:$booting
    vhost_name=$(wds_curl GET "api/v2/sync/block/volumes/$volume_id" | jq -r .volume_detail.name)
    vhost_id=$(wds_curl GET "api/v2/sync/block/vhost?name=$vhost_name" | jq -r '.vhosts[0].id')
    vhost_paths=$(wds_curl GET "api/v2/sync/block/volumes/$volume_id/bind_status" | jq -r .path)
    for k in {1..1}; do
        uss_ret=$(wds_curl PUT "api/v2/sync/block/vhost/unbind_uss" "{\"vhost_id\": \"$vhost_id\", \"uss_gw_id\": \"$src_uss_id\", \"is_snapshot\": false}")
        ret_code=$(echo $uss_ret | jq -r .ret_code)
        if [ "$ret_code" = "0" ]; then
            nvpaths=$(jq length <<< $vhost_paths)
            j=0
            while [ $j -lt $nvpaths ]; do
                vhost_path=$(jq -r .[$j] <<<$vhost_paths)
                wds_curl DELETE "api/v2/failure_domain/black_list" "{\"path\": \"$vhost_path\"}"
                let j=$j+1
            done
            break
        fi
        sleep 1
    done
    let i=$i+1
done
