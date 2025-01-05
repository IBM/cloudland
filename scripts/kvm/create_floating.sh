#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 7 ] && echo "$0 <router> <ext_ip> <ext_gw> <ext_vlan> <int_ip> <int_vlan> <mark_id> <inbound> <outbound>" && exit -1

ID=$1
router=router-$1
ext_cidr=$2
ext_ip=${2%/*}
ext_gw=${3%/*}
ext_vlan=$4
int_addr=$5
int_ip=${int_addr%/*}
int_vlan=$6
mark_id=$(($7 % 4294967295))
inbound=$8
outbound=$9

[ -z "$router" -o "$router" = "router-0" -o  -z "$ext_ip" -o -z "$int_ip" ] && exit 1
ip netns list | grep -q $router
[ $? -ne 0 ] && echo "Router $router does not exist" && exit -1

rtID=$(( $ext_vlan % 250 + 2 ))
table=fip-$ext_vlan
rt_file=/etc/iproute2/rt_tables
grep "^$rtID $table" $rt_file
if [ $? -ne 0 ]; then
    for i in {1..250}; do
        grep "^$rtID\s" $rt_file
	[ $? -ne 0 ] && break
        rtID=$(( ($rtID + 17) % 250 + 2 ))
    done
    echo "$rtID $table" >>$rt_file
fi
suffix=${ID}-${ext_vlan}
ext_dev=te-$suffix
./create_veth.sh $router ext-$suffix te-$suffix
ip link set te-$suffix netns $router 

ip netns exec $router ip addr add $ext_cidr dev $ext_dev
ip netns exec $router ip route add default via $ext_gw table $table
ip_net=$(ipcalc -b $int_addr | grep Network | awk '{print $2}')
ip netns exec $router ip route add $ip_net dev ns-$int_vlan table $table
ip netns exec $router ip rule add from $int_ip lookup $table
ip netns exec $router ip rule add to $int_ip lookup $table
ip netns exec $router iptables -t nat -S | grep " $int_ip\>.* MASQUERADE"
[ $? -ne 0 ] && ip netns exec $router iptables -t nat -I POSTROUTING -s $int_ip -m set ! --match-set nonat dst -j MASQUERADE
ip netns exec $router iptables -t nat -D PREROUTING -d $ext_ip -j DNAT --to-destination $int_ip
ip netns exec $router iptables -t nat -I PREROUTING -d $ext_ip -j DNAT --to-destination $int_ip
ip netns exec $router iptables -t nat -D POSTROUTING -s $int_ip -d $int_ip -j SNAT --to-source $ext_ip
ip netns exec $router iptables -t nat -I POSTROUTING -s $int_ip -d $int_ip -j SNAT --to-source $ext_ip
ip netns exec $router arping -c 3 -I $ext_dev -s $ext_ip $ext_gw

if [ "$inbound" -gt 0 ]; then
    ip netns exec $router iptables -t mangle -D PREROUTING -d $ext_ip -j MARK --set-mark $mark_id
    ip netns exec $router iptables -t mangle -I PREROUTING -d $ext_ip -j MARK --set-mark $mark_id
    ip netns exec $router iptables -D FORWARD -m mark --mark $mark_id -j DROP
    ip netns exec $router iptables -I FORWARD -m mark --mark $mark_id -j DROP
    pkt_rate_limit=$(( $inbound * 100 ))
    pkt_burst_limit=$inbound
    ip netns exec $router iptables -D FORWARD -m mark --mark $mark_id -m limit --limit $pkt_rate_limit/second --limit-burst $pkt_burst_limit -j ACCEPT
    ip netns exec $router iptables -I FORWARD -m mark --mark $mark_id -m limit --limit $pkt_rate_limit/second --limit-burst $pkt_burst_limit -j ACCEPT
    ip netns exec $router tc qdisc add dev ns-$int_vlan root handle 1: htb default 10
    ip netns exec $router tc class add dev ns-$int_vlan parent 1: classid 1:$mark_id htb rate ${bandwidth}mbit burst 100
    ip netns exec $router tc filter add dev ns-$int_vlan protocol ip parent 1:0 prio 1 handle $mark_id fw flowid 1:$mark_id
fi
if [ "$outbound" -gt 0 ]; then
    ip netns exec $router tc filter add dev $ext_dev protocol ip parent 1:0 prio 1 handle 1 fw flowid 1:1
    ip netns exec $router tc class add dev $ext_dev parent 1: classid 1:$mark_id htb rate ${bandwidth}mbit burst 100
    ip netns exec $router tc filter add dev $ext_dev protocol ip parent 1:0 prio 1 u32 match ip dst $ext_ip/32 flowid 1:$mark_id
fi
