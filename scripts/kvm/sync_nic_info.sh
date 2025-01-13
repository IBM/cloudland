#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <vm_ID>" && exit -1

ID=$1
vlans=$(cat)
nvlan=$(jq length <<< $vlans)
i=0
while [ $i -lt $nvlan ]; do
    vlan=$(jq -r .[$i].vlan <<< $vlans)
    ip=$(jq -r .[$i].ip_address <<< $vlans)
    mac=$(jq -r .[$i].mac_address <<< $vlans)
    gateway=$(jq -r .[$i].gateway <<< $vlans)
    router=$(jq -r .[$i].router <<< $vlans)
    inbound=$(jq -r .[$i].inbound <<< $vlans)
    outbound=$(jq -r .[$i].outbound <<< $vlans)
    allow_spoofing=$(jq -r .[$i].allow_spoofing <<< $vlans)
    jq -r .[$i].security <<< $vlans | ./apply_vm_nic.sh "$ID" "$vlan" "$ip" "$mac" "$gateway" "$router" "$inbound" "$outbound" "$allow_spoofing"
    let i=$i+1
done
