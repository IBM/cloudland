#!/bin/bash

cd `dirname $0`
source ../../cloudrc

[ $# -lt 3 ] && echo "$0 <vm_ip> <vm_mac> <allow_spoofing> [nic_name] [vm_ID]" && exit -1

vm_ip=${1%%/*}
vm_mac=$2
allow_spoofing=$3
nic_name=$4
vm_ID=$5

./clear_sg_chain.sh "$nic_name" "true"
../create_sg_chain.sh "$nic_name" "$vm_ip" "$vm_mac" "$allow_spoofing"
../apply_sg_rule.sh "$nic_name"
touch $async_job_dir/$nic_name
