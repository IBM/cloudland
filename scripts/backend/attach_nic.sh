#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && echo "$0 <vm_ID> <vlan> <vm_ip> <vm_mac>" && exit -1

vm_ID=inst-$1
vlan=$2
vm_ip=$3
vm_mac=$4
nic_name=tap$(echo $vm_mac | cut -d: -f4- | tr -d :)
vm_br=br$vlan
[ "$vm_br" = "br$external_vlan" -a -n "$zlayer2_interface" ] && sudo /usr/sbin/bridge fdb add $vm_mac dev $zlayer2_interface
./create_link.sh $vlan
virsh domiflist $vm_ID | grep $vm_mac
if [ $? -ne 0 ]; then
    virsh attach-interface $vm_ID bridge $vm_br --model virtio --mac $vm_mac --target $nic_name --live
    virsh attach-interface $vm_ID bridge $vm_br --model virtio --mac $vm_mac --target $nic_name --config
fi
./create_sg_chain.sh $nic_name $vm_ip $vm_mac
./apply_sg_rule.sh $nic_name
