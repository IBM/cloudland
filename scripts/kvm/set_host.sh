#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 5 ] && echo "$0 <router> <vlan> <mac> <name> <ip> [domain]"

router=router-$1
vlan=$2
vm_mac=$3
vm_name=$4
vm_ip=${5%%/*}
domain=$6
[ -z "$domain" ] && domain=$cloud_domain

vlan_dir=$cache_dir/router/$router/$vlan
mkdir -p $vlan_dir
dhcp_host=$vlan_dir/dhcp_hosts
sed -i "/\<$vm_ip\>/d" $dhcp_host
echo "$vm_mac,$vm_name.$domain,$vm_ip" >> $dhcp_host
dnsmasq_pid=$(ps -ef | grep dnsmasq | grep "\<interface=ns-$vlan\>" | awk '{print $2}')
[ -n "$dnsmasq_pid" ] && kill -HUP $dnsmasq_pid
echo "DHCP config for $vm_mac: $vm_ip in vlan $vlan was setup."
