#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && die "$0 <vlan>"

vlan=$1
br_name=br$vlan
tun_name=tun$vlan
local_ip=$(ip addr show $vlan_interface | grep 192.168.1 | awk '{print $2}' | cut -d/ -f1)

ovs-vsctl br-exists $br_name && exit 0
ovs-vsctl add-br $br_name -- set Bridge $br_name fail-mode=secure
ovs-vsctl --may-exist add-port $br_name $tun_name -- set interface $tun_name type=vxlan options:local_ip=$local_ip options:remote_ip=flow options:key=flow
tun_port=$(ovs-vsctl get Interface $tun_name ofport)
ovs-ofctl add-flow $br_name "table=0,priority=80,arp,actions=set_field:100100->tun_id,set_field:${resolver_addr}->tun_dst,$tun_port"
ovs-ofctl add-flow $br_name "table=0,priority=120,in_port=$tun_port,tun_id=100100,actions=normal,learn(table=10,NXM_OF_ETH_DST[]=NXM_OF_ETH_SRC[],load:NXM_NX_TUN_ID[]->NXM_NX_TUN_ID[],load:NXM_NX_TUN_IPV4_SRC[]->NXM_NX_TUN_IPV4_DST[],output=NXM_OF_IN_PORT[])"
