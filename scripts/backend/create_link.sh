#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 1 ] && die "$0 <vlan> [interface]"

vlan=$1
vm_br=br$vlan
interface=$2

cat /proc/net/dev | grep -q "\<$vm_br\>:"
if [ $? -eq 0 ]; then
    [ "$vlan" = "$external_vlan" -o "$vlan" = "$internal_vlan" ] && exit 0
else
    nmcli connection add con-name $vm_br type bridge ifname $vm_br ipv4.method static ipv4.addresses 169.254.169.254/32
    nmcli connection up $vm_br
fi
cat /proc/net/dev | grep -q "\<v-$vlan\>:"
if [ $? -ne 0 ]; then
    if [ $vlan -ge 4095 ]; then
        [ -z "$interface" ] && interface=$vxlan_interface
        nmcli connection add con-name v-$vlan type vxlan id $vlan ifname v-$vlan remote $vxlan_mcast_addr dev $interface ipv4.method disabled master $vm_br
    else
        [ -z "$interface" ] && interface=$vlan_interface
        nmcli connection add con-name v-$vlan type vlan id $vlan ifname v-$vlan dev $interface ipv4.method disabled master $vm_br
    fi
    nmcli connection up v-$vlan
fi
udevadm settle
