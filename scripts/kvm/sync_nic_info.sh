#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <vm_ID> <hostname>" && exit -1

ID=$1
vm_name=$2
vlans=$(cat)
nvlan=$(jq length <<< $vlans)
i=0
while [ $i -lt $nvlan ]; do
    read -d'\n' -r vlan ip mac gateway router inbound outbound allow_spoofing < <(jq -r ".[$i].vlan, .[$i].ip_address, .[$i].mac_address, .[$i].gateway, .[$i].router, .[$i].inbound, .[$i].outbound, .[$i].allow_spoofing" <<<$vlans)
    jq -r .[$i].security <<< $vlans | ./apply_vm_nic.sh "$ID" "$vlan" "$ip" "$mac" "$gateway" "$router" "$inbound" "$outbound" "$allow_spoofing"
    ./set_host.sh "$router" "$vlan" "$mac" "$vm_name" "$ip"
    jq -r .[$i].sites_ip_info | ./apply_sites_ip.sh "$router"
    let i=$i+1
done
