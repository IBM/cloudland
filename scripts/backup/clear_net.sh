#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <vlan> <network> <tag_id>" && exit -1

vlan=$1
network=$2
tag_id=$3

nspace=vlan$vlan
rm -rf $dmasq_dir/$nspace/tags/$tag_id
dmasq_cmd=$(ps -ef | grep dnsmasq | grep "\<interface=ns-$vlan\>")
dns_pid=$(echo "$dmasq_cmd" | awk '{print $2}')
[ -z "$dns_pid" ] || kill $dns_pid ||  kill -9 $dns_pid
exist_ranges=$(echo "$dmasq_cmd" | tr -s ' ' '\n' | grep "\-\-dhcp-range" | grep -v "set:tag$vlan-$tag_id,")
if [ -n "$exist_ranges" ]; then
    dns_host=$dmasq_dir/$nspace/${nspace}.host
    dns_opt=$dmasq_dir/$nspace/${nspace}.opts
    dns_sh=$dmasq_dir/$nspace/${nspace}.sh
    pid_file=$dmasq_dir/$nspace/${nspace}.pid
    cmd="/usr/sbin/dnsmasq --no-hosts --no-resolv --strict-order --bind-interfaces --interface=ns-$vlan --except-interface=lo --pid-file=$pid_file --dhcp-hostsfile=$dns_host --dhcp-optsfile=$dns_opt --leasefile-ro --dhcp-ignore='tag:!known' $exist_ranges"
    echo "$cmd" > $dns_sh
    chmod +x $dns_sh
    ip netns exec $nspace $dns_sh
else
    ip link del tap-$vlan
    ip netns exec $nspace ip link set lo down
    ip netns del $nspace
    ./clear_link.sh $vlan
    rm -rf $dmasq_dir/$nspace
fi

echo "Network $network was cleared."
