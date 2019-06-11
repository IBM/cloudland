#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && echo "$0 <vm_ID> <vlan> <vm_ip> <vm_mac>" && exit -1

vm_ID=$1
vlan=$2
vm_ip=$3
vm_mac=$4
nic_name=tap$(echo $vm_mac | cut -d: -f4- | tr -d :)
vm_br=br$vlan
./create_link.sh $vlan
virsh attach-interface $vm_ID bridge $vm_br --model virtio --mac $vm_mac --config --target $nic_name
./create_sg_chain.sh $vm_ip $vm_mac
./apply_sg_rule.sh $nic_name
