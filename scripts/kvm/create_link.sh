#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 1 ] && die "$0 <vlan> [interface]"

vlan=$1
vm_br=br$vlan
interface=$2

cat /proc/net/dev | grep -q "\<$vm_br\>:"
[ $? -eq 0 ] && exit 0
nmcli connection add con-name $vm_br type bridge ifname $vm_br ipv4.method static ipv4.addresses 169.254.169.254/32
nmcli connection modify $vm_br bridge.stp no
nmcli connection modify $vm_br bridge.forward-delay 0
nmcli connection up $vm_br
apply_bridge -I $vm_br
cat /proc/net/dev | grep -q "\<v-$vlan\>:"
if [ $? -ne 0 ]; then
    if [ $vlan -ge 4095 ]; then
        [ -z "$interface" ] && interface=$vxlan_interface
        nmcli connection add con-name v-$vlan type vxlan id $vlan vxlan.proxy $proxy_mode ifname v-$vlan dev $interface ipv4.method disabled master $vm_br
    else
        [ -z "$interface" ] && interface=$vlan_interface
        nmcli connection add con-name v-$vlan type vlan id $vlan ifname v-$vlan dev $interface ipv4.method disabled master $vm_br
    fi
    nmcli connection up v-$vlan
fi
udevadm settle
