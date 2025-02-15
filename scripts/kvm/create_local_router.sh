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

def_route=$(ip netns exec router-0 ip route | grep default)
if [ -z "$def_route" ]; then
    echo "|:-COMMAND-:| system_router.sh '$SCI_CLIENT_ID' '$HOSTNAME'"
fi

./create_veth.sh $router int-$suffix ti-$suffix
[ -z "$system_packet_rate_limit" ] && system_packet_rate_limit=120
system_packet_burst=$(( $system_packet_rate_limit / 2 ))
ip netns exec $router iptables -I FORWARD -i ti-$suffix -j DROP
ip netns exec $router iptables -I FORWARD -o ti-$suffix -j DROP
ip netns exec $router iptables -I FORWARD -i ti-$suffix -m limit --limit $system_packet_rate_limit/second --limit-burst $system_packet_burst -j ACCEPT
ip netns exec $router iptables -I FORWARD -o ti-$suffix -m limit --limit $system_packet_rate_limit/second --limit-burst $system_packet_burst -j ACCEPT
remaineder=$(( $suffix % 64516 ))
part2=$(( $remaineder / 254 ))
part3=$(( $remaineder % 254 ))
for i in {1..125}; do
    part4=$(( ($RANDOM % 125) * 2 + 3))
    local_ip=169.$part2.$part3.$part4
    peer_ip=169.$part2.$part3.$(( $part4 - 1 ))
    ip netns exec router-0 ip addr | grep "\<$peer_ip\>"
    [ $? -ne 0 ] && break
done
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
