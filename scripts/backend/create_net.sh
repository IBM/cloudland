#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 6 ] && echo "$0 <vlan> <network> <netmask> <gateway> <dhcp_ip> [tag_id] [role]" && exit -1

vlan=$1
network=$2
netmask=$3
gateway=$4
dhcp_ip=$5
tag_id=$6
role=$7

vm_br=br$vlan
./create_link.sh $vlan

nspace=vlan$vlan
cmdfile=$dmasq_dir/$nspace/tags/$tag_id
mkdir -p $cmdfile
echo "$0 '$vlan' '$network' '$netmask' '$gateway' '$dhcp_ip' '$tag_id' 'BOOT'" > $cmdfile/cmd
if [ ! -f /var/run/netns/$nspace ]; then
    ip netns add $nspace
    ip link add ns-$vlan type veth peer name tap-$vlan
    brctl addif $vm_br tap-$vlan
    ip link set tap-$vlan up
    apply_vnic -I tap-$vlan
    ip link set ns-$vlan netns $nspace
    ip netns exec $nspace ip link set ns-$vlan up
    ip netns exec $nspace ip link set lo up
fi
dns_host=$dmasq_dir/$nspace/${nspace}.host
dns_opt=$dmasq_dir/$nspace/${nspace}.opts
dns_sh=$dmasq_dir/$nspace/${nspace}.sh
pid_file=$dmasq_dir/$nspace/${nspace}.pid
pfix=`ipcalc -p $dhcp_ip | cut -d'=' -f2`
brd=`ipcalc -b $dhcp_ip | cut -d'=' -f2`
ip netns exec $nspace ip addr add $dhcp_ip brd $brd dev ns-$vlan

if [ "$role" != "BOOT" ]; then
    ipcalc -c $gateway >/dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo "tag:tag$vlan-$tag_id,option:router,$gateway" >> $dns_opt
    else
        echo "tag:tag$vlan-$tag_id,option:router" >> $dns_opt
    fi
[ -n "$dns_server" ] && echo "tag:tag$vlan-$tag_id,option:dns-server,$dns_server" >> $dns_opt
fi

dmasq_cmd=$(ps -ef | grep dnsmasq | grep "\<interface=ns-$vlan\>")
dns_pid=$(echo "$dmasq_cmd" | awk '{print $2}')
if [ -z "$dns_pid" ]; then
    cmd="/usr/sbin/dnsmasq --no-hosts --no-resolv --strict-order --bind-interfaces --interface=ns-$vlan --except-interface=lo --pid-file=$pid_file --dhcp-hostsfile=$dns_host --dhcp-optsfile=$dns_opt --leasefile-ro --dhcp-ignore='tag:!known' --dhcp-range=set:tag$vlan-$tag_id,$network,static,86400s"
else
    kill $dns_pid || kill -9 $dns_pid
    exist_ranges=`echo "$dmasq_cmd" | tr -s ' ' '\n' | grep "\-\-dhcp-range"`
    cmd="/usr/sbin/dnsmasq --no-hosts --no-resolv --strict-order --bind-interfaces --interface=ns-$vlan --except-interface=lo --pid-file=$pid_file --dhcp-hostsfile=$dns_host --dhcp-optsfile=$dns_opt --leasefile-ro --dhcp-ignore='tag:!known' --dhcp-range=set:tag$vlan-$tag_id,$network,static,86400s $exist_ranges"
fi
ip netns exec $nspace $cmd
echo "|:-COMMAND-:| $(basename $0) '$vlan' '$SCI_CLIENT_ID' '$role'"
