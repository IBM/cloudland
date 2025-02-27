#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 4 ] && echo "$0 <vm_ID> <volume_ID> <path> <wds_volume_id>" && exit -1

ID=$1
vm_ID=inst-$1
vol_ID=$2
vol_PATH=$3
wds_volume_id=$4

get_wds_token
count=$(virsh dumpxml $vm_ID | grep -c "<disk type='vhostuser' device='disk'")
let letter=97+$count
vhost_name=instance-$ID-vol-$vol_ID-$RANDOM
uss_id=$(get_uss_gateway)
vhost_id=$(wds_curl POST "api/v2/sync/block/vhost" "{\"name\": \"$vhost_name\"}" | jq -r .id)
ret_code=$(wds_curl PUT "api/v2/sync/block/vhost/bind_uss" "{\"vhost_id\": \"$vhost_id\", \"uss_gw_id\": \"$uss_id\", \"lun_id\": \"$wds_volume_id\", \"is_snapshot\": false}" | jq -r .ret_code)
if [ "$ret_code" != "0" ]; then
    wds_curl DELETE "api/v2/sync/block/vhost/$vhost_id"
    echo "|:-COMMAND-:| $(basename $0) '' '$vol_ID' ''"
    exit -1
fi
ux_sock=/var/run/wds/$vhost_name

vol_xml=$xml_dir/$vm_ID/disk-${vol_ID}.xml
cp $template_dir/wds_volume.xml $vol_xml
device=vd$(printf "\\$(printf '%03o' "$letter")")
sed -i "s#VM_UNIX_SOCK#$ux_sock#g;s#VOLUME_TARGET#$device#g;" $vol_xml
timeout_virsh attach-device $vm_ID $vol_xml --config --persistent
if [ $? -eq 0 ]; then
    echo "|:-COMMAND-:| $(basename $0) '$1' '$vol_ID' '$device'"
else
    wds_curl PUT "api/v2/sync/block/vhost/unbind_uss" "{\"vhost_id\": \"$vhost_id\", \"uss_gw_id\": \"$uss_id\", \"is_snapshot\": false}"
    wds_curl DELETE "api/v2/sync/block/vhost/$vhost_id"
    echo "|:-COMMAND-:| $(basename $0) '' '$vol_ID' ''"
fi
vm_xml=$xml_dir/$vm_ID/$vm_ID.xml
timeout_virsh dumpxml --security-info $vm_ID 2>/dev/null | sed "s/autoport='yes'/autoport='no'/g" > $vm_xml.dump && mv -f $vm_xml.dump $vm_xml
