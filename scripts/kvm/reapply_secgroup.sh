#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <vm_ip> <vm_mac> <allow_spoofing>" && exit -1

vm_ip=${1%%/*}
vm_mac=$2
allow_spoofing=$3

nic_name=tap$(echo $vm_mac | cut -d: -f4- | tr -d :)
./clear_sg_chain.sh "$nic_name"
./create_sg_chain.sh "$nic_name" "$vm_ip" "$vm_mac" "$allow_spoofing"
./apply_sg_rule.sh "$nic_name"
