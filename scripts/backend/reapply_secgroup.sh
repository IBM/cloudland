#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <vm_ip> <vm_mac>" && exit -1

vm_ip=${1%%/*}
vm_mac=$2
nic_name=tap$(echo $vm_mac | cut -d: -f4- | tr -d :)
./clear_sg_chain.sh $nic_name
./create_sg_chain.sh $nic_name $vm_ip $vm_mac
./apply_sg_rule.sh $nic_name
