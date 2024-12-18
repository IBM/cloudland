#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <router>" && exit -1

router=$1

[ "${router/router-/}" = "$router" ] && router=router-$1
[ -z "$router" -o "$router" = "router-0" ] && exit 1
[ -f "/var/run/netns/$router" ] && exit 0

ip netns add $router
#ip netns exec $router iptables -A INPUT -m mark --mark 0x1/0xffff -j ACCEPT
ip netns exec $router ip link set lo up
suffix=${router/router-/}

route_ip=$(ip netns exec router-0 ifconfig link-sys | grep 'inet ' | awk '{print $2}')
if [ -z "$route_ip" ]; then
    echo "|:-COMMAND-:| system_router.sh '$SCI_CLIENT_ID' '$HOSTNAME'"
fi

./create_veth.sh $router int-$suffix ti-$suffix
local_ip=169.$(($SCI_CLIENT_ID % 234)).$(($suffix % 234)).3
peer_ip=169.$(($SCI_CLIENT_ID % 234)).$(($suffix % 234)).2
ip netns exec $router ip addr add ${local_ip}/31 dev ti-$suffix
ip netns exec $router ip route add default via $peer_ip

[ ! -f /var/run/netns/router-0 ] && ip netns add router-0
ip link set int-$suffix netns router-0
ip netns exec router-0 ip link set int-$suffix up
ip netns exec router-0 ip addr add ${peer_ip}/31 dev int-$suffix

ip netns exec $router ipset create nonat nethash
ip netns exec $router iptables -t nat -S | grep "to-source $local_ip\>"
[ $? -ne 0 ] && ip netns exec $router iptables -t nat -A POSTROUTING -m set --match-set nonat src -m set ! --match-set nonat dst -j SNAT --to-source $local_ip

ip netns exec $router bash -c "echo 1 >/proc/sys/net/ipv4/ip_forward"
