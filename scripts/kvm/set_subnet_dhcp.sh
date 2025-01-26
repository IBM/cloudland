#!/bin/bash -xv

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <router> <vlan> <gateway> [name_server] [domain_search]" && exit -1

router=$1
[ "${router/router-/}" = "$router" ] && router=router-$1
vlan=$2
gateway_cidr=$3
gateway=${3%%/*}
name_server=$4
domain_search=$5
[ -z "$name_server" ] && name_server=$dns_server
[ -z "$domain_search" ] && domain_search=$cloud_domain

[ "$vlan" -le 4095 ] && exit 0
let mtu=$(ip -o link show $vxlan_interface | cut -d' ' -f5)-50

ipcalc_ret=$(ipcalc -b $gateway_cidr)
network=$(echo "$ipcalc_ret" | grep Network | awk '{print $2}')
netstr=$(echo $network | tr -s './' '_')
hostmin=$(echo "$ipcalc_ret" | grep HostMin | awk '{print $2}')
hostmax=$(echo "$ipcalc_ret" | grep HostMax | awk '{print $2}')
vlan_dir=$cache_dir/router/$router/dnsmasq-$vlan
dnsmasq_conf=$vlan_dir/dnsmasq.conf
dhcp_host=$vlan_dir/dhcp_hosts
dns_host=$vlan_dir/dns_hosts
pid_file=$vlan_dir/dnsmasq.pid
dnsmasq_conf_dir=$vlan_dir/dnsmasq.conf.d
mkdir -p $dnsmasq_conf_dir
cat >$dnsmasq_conf <<EOF
no-hosts
cache-size=0
bind-interfaces
server=$name_server
addn-hosts=$dns_host
strict-order
except-interface=lo
pid-file=$pid_file
log-facility=/var/log/dnsmasq.log
dhcp-hostsfile=$dhcp_host
dhcp-option=26,$mtu
listen-address=$gateway
leasefile-ro
dhcp-ignore=tag:!known
conf-dir=$dnsmasq_conf_dir
EOF

dhcp_conf=$dnsmasq_conf_dir/dhcp_$netstr.conf
cat >$dhcp_conf <<EOF
dhcp-range=$hostmin,$hostmax,2h
dhcp-option=tag:$network,3,$gateway
dhcp-option=tag:$network,6,$gateway,$name_server
dhcp-option=tag:$network,119,$domain_search
EOF

dmasq_cmd=$(ps -ef | grep dnsmasq | grep "\<interface=ns-$vlan\>")
dns_pid=$(echo "$dmasq_cmd" | awk '{print $2}')
if [ -z "$dns_pid" ]; then
    cmd="/usr/sbin/dnsmasq --interface=ns-$vlan -C $dnsmasq_conf"
    ip netns exec $router $cmd
fi
