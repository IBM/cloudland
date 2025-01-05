#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 8 ] && echo "$0 <vm_ID> <vlan> <vm_ip> <vm_mac> <gateway> <router> <inbound> <outbound>" && exit -1

ID=$1
vm_ID=inst-$ID
vlan=$2
vm_ip=$3
vm_mac=$4
gateway=$5
router=$6
inbound=$7
outbound=$8
nic_name=tap$(echo $vm_mac | cut -d: -f4- | tr -d :)
vm_br=br$vlan
./create_link.sh $vlan
brctl setageing $vm_br 0
virsh domiflist $vm_ID | grep $vm_mac
if [ $? -ne 0 ]; then
    virsh attach-interface $vm_ID bridge $vm_br --model virtio --mac $vm_mac --target $nic_name --live
    virsh attach-interface $vm_ID bridge $vm_br --model virtio --mac $vm_mac --target $nic_name --config
fi
./set_nic_speed.sh "$ID" "$nic_name" "$inbound" "$outbound"
./create_sg_chain.sh $nic_name $vm_ip $vm_mac
./apply_sg_rule.sh $nic_name
./set_subnet_gw.sh $router $vlan $gateway $ext_vlan
