#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && die "$0 <router> <vlan> <mac> <ip>"

router=$1
[ "${router/router-/}" = "$router" ] && router=router-$1
vlan=$2
vm_mac=$3
vm_ip=${4%%/*}

vlan_dir=$cache_dir/router/$router/$vlan
dhcp_host=$vlan_dir/dhcp_hosts
sed -i "/^$vm_mac/d" $dhcp_host
sed -i "/\<$vm_ip\>/d" $dhcp_host
ip netns exec $router dhcp_release ns-$vlan $vm_ip $vm_mac
dns_pid=`ps -ef | grep dnsmasq | grep "\<interface=ns-$vlan\>" | awk '{print $2}'`
[ -n "$dns_pid" ] && kill -HUP $dns_pid
ip netns exec $router dhcp_release ns-$vlan $vm_ip $vm_mac
echo "DHCP config for $vm_mac: $vm_ip in vlan $vlan was removed."
