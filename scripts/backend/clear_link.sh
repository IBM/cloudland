#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <vlan>" && exit -1

vlan=$1
vm_br=br$vlan
slaves=$(ls -A /sys/devices/virtual/net/$vm_br/brif | grep -v v-)
[ -n "$slaves" ] && exit 0
nmcli connection down v-$vlan
nmcli connection del v-$vlan
nmcli connection down $vm_br
nmcli connection del $vm_br
