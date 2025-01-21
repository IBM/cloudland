#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 7 ] && echo "$0 <vlan> <ip> <mac> <gateway> <router> <name_server> <domain_search>" && exit -1

vlan=$1
dhcp_ip=$2
mac_address=$3
gateway=${4%%/*}
router=$5
name_server=$6
domain_search=$7
[ -z "$name_server" ] && name_server=$dns_server

dns_host=$router_dir/$vlan/${nspace}.host
dns_opt=$router_dir/$vlan/${nspace}.opts
dns_sh=$router_dir/$vlan/${nspace}.sh
pid_file=$router_dir/$vlan/${nspace}.pid
pfix=$(ipcalc -b $dhcp_ip | grep Netmask | awk '{print $4}')
brd=$(ipcalc -b $dhcp_ip | grep Broadcast | awk '{print $2}')
ip netns exec $nspace ip addr add $dhcp_ip brd $brd dev ns-$vlan

ipcalc -b $gateway >/dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "tag:tag$vlan-$tag_id,option:router,$gateway" >> $dns_opt
else
    echo "tag:tag$vlan-$tag_id,option:router" >> $dns_opt
fi
[ -n "$name_server" ] && echo "tag:tag$vlan-$tag_id,option:dns-server,$name_server" >> $dns_opt
[ -n "$domain_search" ] && echo "tag:tag$vlan-$tag_id,option:domain-search,$domain_search" >> $dns_opt

if [ "$vlan" -gt 4095 ]; then
    let mtu=$(ip -o link show $vxlan_interface | cut -d' ' -f5)-50
    [ "$mtu" -ge 1400 ] && mtu_args="--dhcp-option-force=26,$mtu"
fi

dmasq_cmd=$(ps -ef | grep dnsmasq | grep "\<interface=ns-$vlan\>")
dns_pid=$(echo "$dmasq_cmd" | awk '{print $2}')
if [ -z "$dns_pid" ]; then
    cmd="/usr/sbin/dnsmasq --no-hosts --cache-size=0 --no-resolv --strict-order --interface=ns-$vlan --except-interface=lo --pid-file=$pid_file --dhcp-hostsfile=$dns_host --dhcp-optsfile=$dns_opt $mtu_args --leasefile-ro --dhcp-ignore='tag:!known' --dhcp-range=set:tag$vlan-$tag_id,$network,static,86400s"
else
    kill $dns_pid || kill -9 $dns_pid
    exist_ranges=`echo "$dmasq_cmd" | tr -s ' ' '\n' | grep "\-\-dhcp-range"`
    cmd="/usr/sbin/dnsmasq --no-hosts --cache-size=0 --no-resolv --strict-order --bind-interfaces --interface=ns-$vlan --except-interface=lo --pid-file=$pid_file --dhcp-hostsfile=$dns_host --dhcp-optsfile=$dns_opt $mtu_args  --leasefile-ro --dhcp-ignore='tag:!known' --dhcp-range=set:tag$vlan-$tag_id,$network,static,86400s $exist_ranges"
fi
ip netns exec $nspace $cmd
