#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 3 ] && die "$0 <vm_ID> <router> <boot_volume>"

ID=$1
vm_ID=inst-$ID
router=$2
boot_volume=$3
vm_xml=$(virsh dumpxml $vm_ID)
virsh undefine $vm_ID
cmd="virsh destroy $vm_ID"
result=$(eval "$cmd")
count=$(echo $vm_xml | xmllint --xpath 'count(/domain/devices/interface)' -)
for (( i=1; i <= $count; i++ )); do
    vif_dev=$(echo $vm_xml | xmllint --xpath "string(/domain/devices/interface[$i]/target/@dev)" -)
    br_name=$(echo $vm_xml | xmllint --xpath "string(/domain/devices/interface[$i]/source/@bridge)" -)
    mac_addr=$(echo $vm_xml | xmllint --xpath "string(/domain/devices/interface[$i]/mac/@address)" -)
    if [ "$use_lb" = "false" ]; then
        br_name=br$SCI_CLIENT_ID
        result=$(icp-tower --ovs-bridge=$br_name gate remove --interface $vif_dev)
    else
        vni=${br_name#br}
        ./clear_sg_chain.sh $vif_dev >/dev/null 2>&1
    fi
done
./clear_local_router.sh $router

rm -f ${cache_dir}/meta/${vm_ID}.iso
rm -rf $xml_dir/$vm_ID
if [ -z "$wds_address" ]; then	
    rm -f ${image_dir}/${vm_ID}.*
else
    get_wds_token
    vhosts=$(basename $(ls /var/run/wds/instance-${ID}-*))
    for vhost_name in $vhosts; do
        if [ -S "/var/run/wds/$vhost_name" ]; then
           vhost_id=$(wds_curl GET "api/v2/sync/block/vhost?name=$vhost_name" | jq -r '.vhosts[0].id')
           uss_id=$(get_uss_gateway)
           wds_curl PUT "api/v2/sync/block/vhost/unbind_uss" "{\"vhost_id\": \"$vhost_id\", \"uss_gw_id\": \"$uss_id\", \"is_snapshot\": false}"
        fi
    done
fi
echo "|:-COMMAND-:| $(basename $0) '$ID'"
