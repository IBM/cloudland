#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 1 ] && die "$0 <vm_ID> <router>"

ID=$1
vm_ID=inst-$1
router=$2
vm_xml=$(virsh dumpxml $vm_ID)
virsh undefine $vm_ID
cmd="virsh destroy $vm_ID"
#sidecar span log $span "Internal: $cmd" "result: $result"
result=$(eval "$cmd")
count=$(echo $vm_xml | xmllint --xpath 'count(/domain/devices/interface)' -)
for (( i=1; i <= $count; i++ )); do
    vif_dev=$(echo $vm_xml | xmllint --xpath "string(/domain/devices/interface[$i]/target/@dev)" -)
    br_name=$(echo $vm_xml | xmllint --xpath "string(/domain/devices/interface[$i]/source/@bridge)" -)
    mac_addr=$(echo $vm_xml | xmllint --xpath "string(/domain/devices/interface[$i]/mac/@address)" -)
    if [ "$use_lb" = "false" ]; then
        br_name=br$SCI_CLIENT_ID
        result=$(icp-tower --ovs-bridge=$br_name gate remove --interface $vif_dev)
        sidecar span log $span "Internal: $vif_dev is deleted" "result: $result"
    else
        vni=${br_name#br}
        ./clear_link.sh $vni
        ./rm_fdb.sh $mac_addr
        ./clear_sg_chain.sh $vif_dev
	./clear_local_router $router
    fi
    sidecar span log $span "Callback: clear_vnic.sh '$vif_dev'"
done

rm -f ${cache_dir}/meta/${vm_ID}.iso
rm -rf $xml_dir/$vm_ID
if [ -z "$wds_address" ]; then	
    rm -f ${image_dir}/${vm_ID}.*
else
    vhost_name=instance-$ID-boot
    vhost_id=$(wds_curl GET "api/v2/sync/block/vhost" | jq --arg vhost $vhost_name -r '.vhosts | .[] | select(.name == $vhost) | .id')
    uss_id=$(wds_curl GET "api/v2/wds/uss" | jq --arg hname $(hostname -s) -r '.uss_gateways | .[] | select(.server_name == $hname) | .id')
    wds_curl PUT "api/v2/sync/block/vhost/unbind_uss" "{\"vhost_id\": \"$vhost_id\", \"uss_gw_id\": \"$uss_id\", \"is_snapshot\": false}"
    wds_curl DELETE "api/v2/sync/block/vhost/$vhost_id"
    volume_id=$(wds_curl GET "api/v2/sync/block/volumes" | jq --arg volume $vhost_name -r '.volumes | .[] | select(.name == $volume) | .id')
    wds_curl DELETE "api/v2/sync/block/volumes/$volume_id?force=false"
fi
sidecar span log $span "Callback: `basename $0` '$vm_ID'"
echo "|:-COMMAND-:| $(basename $0) '$ID'"
