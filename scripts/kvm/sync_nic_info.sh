#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <vm_ID> <hostname>" && exit -1

ID=$1
vm_name=$2
domain_search=$cloud_domain
if [ "${vm_name%%.*}" != "$vm_name" ]; then
    domain_search=$(echo $vm_name | cut -d. -f2-)
fi
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
    ./set_subnet_gw.sh "$router" "$vlan" "$gateway"
    ./set_subnet_dhcp.sh "$router" "$vlan" "$gateway" "$dns_server" "$domain_search"
    let i=$i+1
done
