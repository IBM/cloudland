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

./create_veth.sh $router ext-$ext_vlan link-$ext_vlan
ip netns exec $router ip addr add $ext_ip dev link-$ext_vlan
ip netns exec $router ip route add default via $gateway
route_ip=${ext_ip%/*}
ip netns exec $router iptables -P INPUT DROP
ip netns exec $router iptables -D FORWARD -p tcp --dport 25 -j DROP
ip netns exec $router iptables -I FORWARD -p tcp --dport 25 -j DROP
ip netns exec $router iptables -D FORWARD -p tcp --dport 465 -j DROP
ip netns exec $router iptables -I FORWARD -p tcp --dport 465 -j DROP
ip netns exec $router iptables -D FORWARD -p tcp --dport 587 -j DROP
ip netns exec $router iptables -I FORWARD -p tcp --dport 587 -j DROP
ip netns exec $router iptables -t nat -S | grep "to-source $ext_ip\>"
[ $? -ne 0 ] && ip netns exec $router iptables -t nat -A POSTROUTING -j SNAT --to-source $route_ip

ip netns exec $router bash -c "echo 1 >/proc/sys/net/ipv4/ip_forward"
