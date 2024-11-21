#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <ext_vlan> <ext_ip> <gateway>" && exit -1

ext_vlan=$1
ext_ip=$2
gateway=${3%/*}
router=router-0

ip netns add $router
ip netns exec $router ip link set lo up

./create_veth.sh $router ext-sys link-sys $ext_vlan
ip netns exec $router ip addr add $ext_ip dev link-sys
ip netns exec $router ip route add default via $gateway
route_ip=${ext_ip%/*}
ip netns exec $router iptables -t nat -S | grep "source $ext_ip\>"
[ $? -ne 0 ] && ip netns exec $router iptables -t nat -A POSTROUTING -j SNAT --to-source $ext_ip

router_dir=$cache_dir/router/$router
mkdir -p $router_dir
ip netns exec $router iptables-save > $router_dir/iptables.save
ip netns exec $router ipset save > $router_dir/ipset.save
ip netns exec $router bash -c "echo 1 >/proc/sys/net/ipv4/ip_forward"
