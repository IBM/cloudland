#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 1 ] && die "$0 <vm_ID>"

vm_ID=inst-$1
vm_xml=$(virsh dumpxml $vm_ID)
virsh undefine $vm_ID
cmd="virsh destroy $vm_ID"
#sidecar span log $span "Internal: $cmd" "result: $result"
result=$(eval "$cmd")
count=$(echo $vm_xml | xmllint --xpath 'count(/domain/devices/interface)' -)
for (( i=1; i <= $count; i++ )); do
    vif_dev=$(echo $vm_xml | xmllint --xpath "string(/domain/devices/interface[$i]/target/@dev)" -)
    br_name=$(echo $vm_xml | xmllint --xpath "string(/domain/devices/interface[$i]/source/@bridge)" -)
    if [ "$use_lb" = "false" ]; then
        br_name=br$SCI_CLIENT_ID
        result=$(icp-tower --ovs-bridge=$br_name gate remove --interface $vif_dev)
        sidecar span log $span "Internal: $vif_dev is deleted" "result: $result"
    else
        vni=${br_name#br}
        ./clear_link.sh $vni
        ./clear_sg_chain.sh $vif_dev
    fi
    sidecar span log $span "Callback: clear_vnic.sh '$vif_dev'"
done

rm -f ${image_dir}/${vm_ID}.*
rm -f ${cache_dir}/meta/${vm_ID}.iso
rm -rf $xml_dir/$vm_ID
sidecar span log $span "Callback: `basename $0` '$vm_ID'"
echo "|:-COMMAND-:| $(basename $0) '$1'"
