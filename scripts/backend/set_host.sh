#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && echo "$0 <vlan> <mac> <name> <ip> [domain]"

vlan=$1
vm_mac=$2
vm_name=$3
vm_ip=${4%%/*}
domain=$5
[ -z "$domain" ] && domain=cloud_domain

nspace=vlan$vlan
dns_host=$dmasq_dir/$nspace/${nspace}.host
sed -i "/\<$vm_ip\>/d" $dns_host
echo "$vm_mac,$vm_name.$domain,$vm_ip" >> $dns_host
dns_pid=$(ps -ef | grep dnsmasq | grep "\<interface=ns-$vlan\>" | awk '{print $2}')
[ -n "$dns_pid" ] && kill -HUP $dns_pid
echo "DHCP config for $vm_mac: $vm_ip in vlan $vlan was setup."
