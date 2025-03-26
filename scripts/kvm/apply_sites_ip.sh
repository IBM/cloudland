#!/bin/bash -xv

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <router> <int_ip>" && exit -1

ID=$1
router=router-$ID
int_ip=${2%/*}
sites=$(cat)
nsite=$(jq length <<< $sites)
i=0
while [ $i -lt $nsite ]; do
    read -d'\n' -r site_vlan gateway < <(jq -r ".[$i].vlan, .[$i].gateway" <<<$sites)
    site_addrs=$(jq -r ".[$i].addresses" <<<$sites)
    gateway=${gateway%/*}
    ./create_route_table.sh $ID $site_vlan
    naddr=$(jq length <<<$site_addrs)
    j=0
    while [ $j -lt $naddr ]; do
        read -d'\n' -r queue_id address < <(jq -r ".[$j].id, .[$j].address" <<<$site_addrs)
        ext_dev=te-${ID}-${site_vlan}
        queue_id=$(($queue_id % 4294967295))
	ext_ip=${address%/*}
        ip netns exec $router ip addr add $address dev $ext_dev
        ip netns exec $router arping -c 3 -U -I $ext_dev $ext_ip
	gw_mac=$(ip netns exec $router arping -c 1 -I $ext_dev $gateway | grep 'Unicast reply' | awk '{print $5}' | tr -d '[]')
	ip netns exec $router iptables -t mangle -D PREROUTING -d $ext_ip/32 -j TEE --gateway $int_ip
	ip netns exec $router iptables -t mangle -I PREROUTING -d $ext_ip/32 -j TEE --gateway $int_ip
	ip netns exec $router iptables -D INPUT -d $ext_ip/32 -j DROP
	ip netns exec $router iptables -I INPUT -d $ext_ip/32 -j DROP
	ip netns exec $router iptables -t mangle -D PREROUTING -s $ext_ip -j NFQUEUE --queue-num $queue_id
	ip netns exec $router iptables -t mangle -I PREROUTING -s $ext_ip -j NFQUEUE --queue-num $queue_id
	ip netns exec $router setsid python3 ./forward_pkt.py "$queue_id" "$ext_dev" "$gw_mac" &
        let j=$j+1
    done
    let i=$i+1
done
