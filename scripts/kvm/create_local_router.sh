#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <router>" && exit -1

ID=$1
router=router-$ID

[ -z "$router" ] && exit 1

ip netns add $router
#ip netns exec $router iptables -A INPUT -m mark --mark 0x1/0xffff -j ACCEPT
ip netns exec $router ip link set lo up
suffix=$1

./create_veth.sh $router ext-$suffix te-$suffix
./create_veth.sh $router int-$suffix ti-$suffix
nat_ip=169.$(($SCI_CLIENT_ID % 234)).$(($suffix % 234)).3
peer_ip=169.$(($SCI_CLIENT_ID % 234)).$(($suffix % 234)).2
ip netns exec $router ip addr add ${nat_ip}/31 dev ti-$suffix
ip addr add ${peer_ip}/31 dev int-$suffix

ip netns exec $router ipset create nonat nethash
ip netns exec $router iptables -t nat -S | grep "source \<$nat_ip\>"
if [ $? -ne 0 ]; then
    ip netns exec $router iptables -t nat -A POSTROUTING -m set --match-set nonat src -m set ! --match-set nonat dst -j SNAT --to-source $nat_ip
    route_ip=$(ifconfig $vxlan_interface | grep 'inet ' | awk '{print $2}')
    iptables -t nat -A POSTROUTING -s ${nat_ip}/32 -j SNAT --to-source $route_ip
    apply_vnic -I ext-$suffix
    apply_vnic -I int-$suffix
fi

router_dir=$cache_dir/router/$router
mkdir -p $router_dir
ip netns exec $router iptables-save > $router_dir/iptables.save
ip netns exec $router ipset save > $router_dir/ipset.save
ip netns exec $router bash -c "echo 1 >/proc/sys/net/ipv4/ip_forward"
