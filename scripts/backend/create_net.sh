#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 6 ] && echo "$0 <vlan> <network> <netmask> <gateway> <start_ip> <end_ip> [tag_id] [role]" && exit -1

vlan=$1
network=$2
netmask=$3
gateway=$4
start_ip=$5
end_ip=$6
tag_id=$7
role=$8

vm_br=br$vlan
./create_link.sh $vlan

ip netns add vlan$vlan
ip link add ns-$vlan type veth peer name tap-$vlan
brctl addif $vm_br tap-$vlan
ip link set tap-$vlan up
ip link set ns-$vlan netns vlan$vlan
ip netns exec vlan$vlan ip link set ns-$vlan up
ip netns exec vlan$vlan ip link set lo up
pfix=`ipcalc -p $start_ip $netmask | cut -d'=' -f2`
brd=`ipcalc -b $start_ip $netmask | cut -d'=' -f2`
ip netns exec vlan$vlan ip addr add $start_ip/$pfix brd $brd dev ns-$vlan

dns_host=$dmasq_dir/vlan$vlan.host
dns_opt=$dmasq_dir/vlan$vlan.opts

dmasq_cmd=`ps -ef | grep dnsmasq | grep "\<interface=ns-$vlan\>" | awk '{print $2}'`
dns_pid=`echo "$dmasq_cmd" | awk '{print $2}'`
if [ -z "$dns_pid" ]; then
    pid_file=$dmasq_dir/vlan$vlan.pid
    ip netns exec vlan$vlan /usr/sbin/dnsmasq --no-hosts --no-resolv --strict-order --bind-interfaces --interface=ns-$vlan --except-interface=lo --pid-file=$pid_file --dhcp-hostsfile=$dns_host --dhcp-optsfile=$dns_opt --leasefile-ro --dhcp-ignore='tag:!known' --dhcp-range=set:tag$vlan-$tag_id,$network,static,86400s
else
    kill $dns_pid || kill -9 $dns_pid
    exist_ranges=`echo "$dmasq_cmd" | tr -s ' ' '\n' | grep "\-\-dhcp-range"`
    ip netns exec vlan$vlan /usr/sbin/dnsmasq --no-hosts --no-resolv --strict-order --bind-interfaces --interface=ns-$vlan --except-interface=lo --pid-file=$pid_file --dhcp-hostsfile=$dns_host --dhcp-optsfile=$dns_opt --leasefile-ro --dhcp-ignore='tag:!known' --dhcp-range=set:tag$vlan-$tag_id,$network,static,86400s $exist_ranges
fi
echo "|:-COMMAND-:| $(basename $0) '$vlan' '$SCI_CLIENT_ID' '$role'"
