#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <vlan> <network> <tag_id>" && exit -1

vlan=$1
network=$2

dmasq_cmd=`ps -ef | grep dnsmasq | grep "\<interface=ns-$vlan\>" | awk '{print $2}'`
dns_pid=`echo "$dmasq_cmd" | awk '{print $2}'`
[ -z "$dns_pid" ] || kill $dns_pid ||  kill -9 $dns_pid
exist_ranges=`echo "$dmasq_cmd" | tr -s ' ' '\n' | grep "\-\-dhcp-range"` | grep -v "set:tag$vlan-$tag_id,"
if [ -n "$exist_ranges" ]; then
    pid_file=$dmasq_dir/vlan$vlan.pid
    dns_host=$dmasq_dir/vlan$vlan.host
    dns_opt=$dmasq_dir/vlan$vlan.opts
    ip netns exec vlan$vlan /usr/sbin/dnsmasq --no-hosts --no-resolv --strict-order --bind-interfaces --interface=ns-$vlan --except-interface=lo --pid-file=$pid_file --dhcp-hostsfile=$dns_host --dhcp-optsfile=$dns_opt --leasefile-ro --dhcp-ignore='tag:!known' $exist_ranges
fi

echo "Network $network was cleared."
