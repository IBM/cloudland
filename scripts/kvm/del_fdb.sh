#!/bin/bash

cd `dirname $0`
source ../cloudrc

exec >>/tmp/del_fdb.log 2>&1
[ $# -lt 5 ] && die "$0 <instance> <vlan> <ip> <mac> <hyper_ip>"

instance=$1
vlan=$2
ip=$3
mac=$4
hyper_ip=$5

grep -q "\<vx-$vlan\>" /proc/net/dev
if [ $? -eq 0 ]; then
    bridge fdb del $mac dev vx-$vlan
#    arp -d $ip dev vx-$vlan
else
    br_name=br$SCI_CLIENT_ID
    tunip=$(inet_aton 172.250.0.10)
    let tunip=$tunip+$SCI_CLIENT_ID
    tunip=$(inet_ntoa $tunip)
#    local_ip=$(ifconfig $vxlan_interface | grep 'inet addr:' | cut -d: -f2 | cut -d' ' -f1)
    cmd="echo $vlan $hyper_ip $ip VXLAN:4789 | icp-tower --debug --ovs-bridge=$br_name route remove --local-ip=$tunip"
    sidecar span log $span "Internal: $cmd" "Result: $result"
    result=$(eval "$cmd")
fi
#sql_exec "delete from vtep where instance='$instance'"
