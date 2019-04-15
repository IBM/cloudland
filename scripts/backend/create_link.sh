#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && die "$0 <vlan>"

vlan=$1
vm_br=br$vlan

cat /proc/net/dev | grep -q "\<$vm_br\>:"
[ $? -eq 0 ] && exit 1

nmcli connection add con-name $vm_br type bridge ifname $vm_br ipv4.method disabled
nmcli connection up $vm_br
if [ $vlan -ge 4095 ]; then
    nmcli connection add con-name v-$vlan type vxlan id $vlan ifname v-$vlan remote $vxlan_mcast_addr dev $vxlan_interface ipv4.method disabled master $vm_br
else
    nmcli connection add con-name v-$vlan type vlan id $vlan ifname v-$vlan dev $vlan_interface ipv4.method disabled master $vm_br
fi
nmcli connection up v-$vlan
