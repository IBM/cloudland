#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <vlan> [force (yes | no)]" && exit -1

vlan=$1
force=$2
[ -z "$force" ] && force=no
vm_br=br$vlan
slaves=$(ls -A /sys/devices/virtual/net/$vm_br/brif | grep -v v-)
[ -n "$slaves" ] && exit 0
if [ "$force" = "no" ]; then
    [ "$vlan" = "$external_vlan" -o "$vlan" = "$internal_vlan" ] && exit 0
fi
nmcli connection down v-$vlan
nmcli connection del v-$vlan
nmcli connection down $vm_br
nmcli connection del $vm_br
