#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 6 ] && die "$0 <vif_name> <vlan> <vm_ip_with_prefix> <vm_mac> <tunip> <decapper>"

vif_name=$1
vlan=$2
vm_ip=$3
vm_mac=$4
tunip=$5
decapper=$6

echo "|:-COMMAND-:| create_vnic.sh '$vif_name' '$SCI_CLIENT_ID' '$vlan' '$vm_ip' '$vm_mac' '$tunip'"
br_name=br$SCI_CLIENT_ID
cmd="icp-tower --ovs-bridge=$br_name gate add --direct-routing --encap-identifier $vlan --local-ip=$tunip --interface $vif_name --vsi-mac-address $vm_mac --vsi-ip-prefix ${vm_ip} --decapper-ip $decapper"
result=$(eval "$cmd")
sidecar span log $span "Internal: $cmd" "Result: $result"
