#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <router> <int_ip>" && exit -1

ID=$1
router=router-$ID
int_ip=${2%/*}
sites=$(cat)
nsite=$(jq length <<< $sites)
i=0
while [ $i -lt $nsite ]; do
    read -d'\n' -r site_vlan gateway < <(jq -r ".[$i].site_vlan, .[$i].gateway" <<<$sites)
    site_addrs=$(jq -r ".[$i].addresses" <<<$sites)
    naddr=$(jq length <<<$site_addrs)
    j=0
    while [ $j -lt $naddr ]; do
        read -d'\n' -r queue_id address < <(jq -r ".[$i].id, .[$i].address" <<<$site_addrs)
        suffix=${ID}-${site_vlan}
        ext_dev=te-$suffix
        ./create_veth.sh $router ext-$suffix te-$suffix
        ip netns exec $router ip addr add $address dev $ext_dev
        ip netns exec $router arping -c 3 -U -I $ext_dev $gateway
	ext_ip=${address%/*}
	ip netns exec $router iptables -t mangle -I PREROUTING -d $ext_ip/32 -j TEE --gateway $int_ip
	ip netns exec $router iptables -I INPUT -d $ext_ip/32 -j DROP
	ip netns exec $router iptables -t mangle -I PREROUTING -s $ext_ip -j NFQUEUE --queue-num $queue_id
	./forward_pkt.py $queue_id $ext_dev
        let j=$j+1
    done
    let i=$i+1
done
