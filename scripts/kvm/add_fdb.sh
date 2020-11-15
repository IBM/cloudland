#!/bin/bash

cd `dirname $0`
source ../cloudrc

exec >>/tmp/add_fdb.log 2>&1
[ $# -lt 5 ] && [ $# -gt 1 ] && die "$0 <instance> <vlan> <hyper_ip> <ip> <mac>"

instance=$1
vlan=$2
hyper_ip=$3
ip=$4
mac=$5

grep -q "\<vx-$vlan\>" /proc/net/dev
if [ $? -ne 0 ]; then
    br_name=br$SCI_CLIENT_ID
	[ -z "$tunip" ] && get_tunip
    if [ $# -eq 5 ]; then
        cmd="echo $vlan $hyper_ip $ip VXLAN:4789 | icp-tower --debug --ovs-bridge=$br_name route add --local-ip=$tunip"
        sidecar span log $span "Internal: $cmd" "Result: $result"
        result=$(eval "$cmd")
    else
        sidecar span log $span "Internal: icp-tower --debug --ovs-bridge=$br_name route add --local-ip=$tunip with stdin" "Result: $result"
        cut -d' ' -f2-4 | sed "/^$/d;s/\(.*\)$/\1 VXLAN:4789/g" | icp-tower --debug --ovs-bridge=$br_name route add --local-ip=$tunip
        sidecar span log $span "Result: $result"
    fi
fi
#    bridge fdb add $mac dev vx-$vlan dst $hyper_ip  self permanent
#    arp -s $ip $mac dev vx-$vlan
#sql_exec "insert into vtep (instance, vni, inner_ip, inner_mac, outer_ip) values ('$instance', '$vlan', '$ip', '$mac', '$hyper_ip')"
