#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 5 ] && die "$0 <instance> <vlan> <ip> <netmask> <mac>"

instance=$1
vlan=$2
ip=$3
netmask=$4
mac=$5

ip netns add $instance
./create_ovsbr.sh $vlan
ip link add dev ${instance}-vs type veth peer name ${instance}-vm
ip link set ${instance}-vs up
br_name=br$vlan
ovs-vsctl add-port $br_name ${instance}-vs
vm_port=$(ovs-vsctl get Interface ${instance}-vs ofport)
ovs-ofctl add-flow $br_name "table=0,priority=100,arp,nw_dst=$ip,actions=output:$vm_port"
ovs-ofctl add-flow $br_name "table=0,priority=100,ip,nw_dst=$ip,actions=output:$vm_port"
ovs-ofctl add-flow $br_name "table=0,priority=100,ip,in_port=$vm_port,actions=resubmit(,10)"
ip link set ${instance}-vm netns $instance
ip netns exec $instance ip link set ${instance}-vm up
prefix=$(ipcalc -p $ip $netmask | cut -d= -f2)
brdcast=$(ipcalc -b $ip $netmask | cut -d= -f2)
ip netns exec $instance ip link set address $mac dev ${instance}-vm
ip netns exec $instance ip link set ${instance}-vm mtu 1450
ip netns exec $instance ip addr add ${ip}/$prefix brd $brdcast dev ${instance}-vm
ip netns exec $instance ip link set lo up
hyper_ip=$(ip addr show $vxlan_interface | grep 192.168.1 | awk '{print $2}' | cut -d/ -f1)
echo "|:-COMMAND-:| /opt/cloudland/scripts/frontback/$(basename $0) '$instance' '$SCI_CLIENT_ID' '$hyper_ip'"
