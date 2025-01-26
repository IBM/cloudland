#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 5 ] && echo "$0 <router> <vlan> <mac> <name> <ip>"

router=router-$1
vlan=$2
vm_mac=$3
vm_name=$4
[ "${vm_name%%.*}" == "$vm_name" ] && vm_name=$vm_name.$cloud_domain
vm_ip=${5%%/*}

vlan_dir=$cache_dir/router/$router/dnsmasq-$vlan
mkdir -p $vlan_dir
dhcp_host=$vlan_dir/dhcp_hosts
sed -i "/\<$vm_ip\>/d" $dhcp_host
echo "$vm_mac,$vm_ip,$vm_name" >> $dhcp_host
for vdir in $cache_dir/router/$router/dnsmasq-*; do
    dns_host=$vdir/dns_hosts
    sed -i "/\<$vm_ip\>/d" $dns_host
    echo "$vm_ip $vm_name" >> $dns_host
done
dnsmasq_pid=$(ps -ef | grep dnsmasq | grep "\<interface=ns-$vlan\>" | awk '{print $2}')
[ -n "$dnsmasq_pid" ] && kill -HUP $dnsmasq_pid
echo "DHCP config for $vm_mac: $vm_ip in vlan $vlan was setup."
