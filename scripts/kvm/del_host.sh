#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && die "$0 <vlan> <mac> <ip>"

vlan=$1
vm_mac=$2
vm_ip=${3%%/*}

nspace=vlan$vlan
dns_host=$dmasq_dir/$nspace/${nspace}.host
sed -i "/^$vm_mac/d" $dns_host
sed -i "/\<$vm_ip\>/d" $dns_host
ip netns exec $nspace dhcp_release ns-$vlan $vm_ip $vm_mac
dns_pid=`ps -ef | grep dnsmasq | grep "\<interface=ns-$vlan\>" | awk '{print $2}'`
[ -n "$dns_pid" ] && kill -HUP $dns_pid
ip netns exec $nspace dhcp_release ns-$vlan $vm_ip $vm_mac
echo "DHCP config for $vm_mac: $vm_ip in vlan $vlan was removed."
