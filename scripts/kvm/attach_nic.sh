#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 6 ] && echo "$0 <vm_ID> <vlan> <vm_ip> <vm_mac> <gateway> <router>" && exit -1

vm_ID=inst-$1
vlan=$2
vm_ip=$3
vm_mac=$4
gateway=$5
router=$6
nic_name=tap$(echo $vm_mac | cut -d: -f4- | tr -d :)
vm_br=br$vlan
./create_link.sh $vlan
brctl setageing $vm_br 0
virsh domiflist $vm_ID | grep $vm_mac
if [ $? -ne 0 ]; then
    virsh attach-interface $vm_ID bridge $vm_br --model virtio --mac $vm_mac --target $nic_name --live
    virsh attach-interface $vm_ID bridge $vm_br --model virtio --mac $vm_mac --target $nic_name --config
fi
./create_sg_chain.sh $nic_name $vm_ip $vm_mac
./apply_sg_rule.sh $nic_name
./create_local_router.sh $router
bcast=$(ipcalc -b $gateway | grep Broadcast | awk '{print $2}')
./create_veth.sh router-$router ln-$vlan ns-$vlan
brctl addif br$vlan ln-$vlan
ipnet=$(ipcalc -b $gateway | grep Network | awk '{print $2}')
ip netns exec $router ipset add nonat $ip_net
ip netns exec router-$router ip addr add $gateway brd $bcast dev ns-$vlan
if [ $? -eq 0 ]; then
    ip netns exec router-$router ip link set ns-$vlan address 52:54:00:00:00:01
fi
