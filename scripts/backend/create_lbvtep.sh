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
./create_link.sh $vlan
bridge fdb add $mac dev vx-$vlan dst 127.0.0.1 self permanent
ip link add dev ${instance}-vs type veth peer name ${instance}-vm
ip link set ${instance}-vs up
ip link set dev ${instance}-vs master br$vlan
ip link set ${instance}-vm netns $instance
ip netns exec $instance ip link set ${instance}-vm up
prefix=$(ipcalc -p $ip $netmask | cut -d= -f2)
brdcast=$(ipcalc -b $ip $netmask | cut -d= -f2)
ip netns exec $instance ip link set address $mac dev ${instance}-vm
ip netns exec $instance ip link set ${instance}-vm mtu 1450
ip netns exec $instance ip addr add ${ip}/$prefix brd $brdcast dev ${instance}-vm
ip netns exec $instance ip link set lo up
hyper_ip=$(ip addr show $vxlan_interface | grep 192.168 | awk '{print $2}' | cut -d/ -f1)
sql_exec "insert into vtep (instance, vni, inner_ip, inner_mac, outer_ip) values ('$instance', '$vlan', '$ip', '$mac', '127.0.0.1')"
echo "|:-COMMAND-:| /opt/cloudland/scripts/frontback/$(basename $0) '$instance' '$SCI_CLIENT_ID' '$hyper_ip'"
