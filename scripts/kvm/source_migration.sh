#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 6 ] && die "$0 <migration_ID> <task_ID> <vm_ID> <router> <target_hyper> <migration_type>"

migration_ID=$1
task_ID=$2
ID=$3
vm_ID=inst-$ID
router=$4
target_hyper=$5
migration_type=$6
state=error

vm_xml=$(virsh dumpxml $vm_ID)
virsh undefine $vm_ID
if [ "$migration_type" = "warm" ]; then
    virsh migrate --live $vm_ID qemu+ssh://$target_hyper/system
    if [ $? -ne 0 ]; then
	state="failed"
        echo "|:-COMMAND-:| migrate_vm.sh '$migration_ID' '$task_ID' '$ID' '$SCI_CLIENT_ID' '$state'"
	exit 1
    fi
else
    virsh shutdown $vm_ID
    for i in {1..60}; do
        vm_state=$(virsh dominfo $vm_ID | grep State | cut -d: -f2- | xargs | sed 's/shut off/shut_off/g')
        [ "$vm_state" = "shut_off" ] && break
        sleep 0.5
    done
    if [ "$vm_state" != "shut_off" ]; then
        virsh destroy $vm_ID
    fi
fi

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
        ./clear_sg_chain.sh $vif_dev
    fi
done
./clear_local_router.sh $router
rm -f ${cache_dir}/meta/${vm_ID}.iso
rm -rf $xml_dir/$vm_ID
state="source_prepared"
echo "|:-COMMAND-:| migrate_vm.sh '$migration_ID' '$task_ID' '$ID' '$SCI_CLIENT_ID' '$state'"
