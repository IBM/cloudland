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
./clear_link.sh $vlan
state=$(virsh dominfo $vm_ID | grep State | cut -d: -f2- | xargs)
if [ "$state" = "running" ]; then
    virsh detach-interface $vm_ID bridge --mac $vm_mac --live --config
else
    virsh detach-interface $vm_ID bridge --mac $vm_mac --config
fi
./clear_sg_chain.sh $nic_name
