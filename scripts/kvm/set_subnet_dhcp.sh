#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 6 ] && echo "$0 <router> <vlan> <gateway> <network> <hostmin> <hostmax> [name_server]" && exit -1

router=$1
[ "${router/router-/}" = "$router" ] && router=router-$1
vlan=$2
gateway=${3%%/*}
network=$4
hostmin=$5
hostmax=$6
name_server=$7
[ -z "$name_server" ] && name_server=$dns_server

[ "$vlan" -le 4095 ] && exit 0
let mtu=$(ip -o link show $vxlan_interface | cut -d' ' -f5)-50

netstr=$(echo $network | tr -s './' '_')
vlan_dir=$cache_dir/router/$router/$vlan
dnsmasq_conf=$vlan_dir/dnsmasq.conf
dhcp_host=$vlan_dir/dhcp_hosts
pid_file=$vlan_dir/dnsmasq.pid
dnsmasq_conf_dir=$vlan_dir/dnsmasq.conf.d
mkdir -p $dnsmasq_conf_dir
cat >$dnsmasq_conf <<EOF
no-hosts
cache-size=0
no-resolv
strict-order
except-interface=lo
pid-file=$pid_file
log-facility=/var/log/dnsmasq.log
dhcp-hostsfile=$dhcp_host
dhcp-option=26,$mtu
leasefile-ro
dhcp-ignore=tag:!known
conf-dir=$dnsmasq_conf_dir
EOF

dhcp_conf=$dnsmasq_conf_dir/dhcp_$netstr.conf
cat >$dhcp_conf <<EOF
dhcp-range=$hostmin,$hostmax,2h
dhcp-option=tag:$network,3,$gateway
dhcp-option=tag:$network,6,$name_server
EOF

dmasq_cmd=$(ps -ef | grep dnsmasq | grep "\<interface=ns-$vlan\>")
dns_pid=$(echo "$dmasq_cmd" | awk '{print $2}')
if [ -z "$dns_pid" ]; then
    cmd="/usr/sbin/dnsmasq --interface=ns-$vlan -C $dnsmasq_conf"
    ip netns exec $router $cmd
fi
