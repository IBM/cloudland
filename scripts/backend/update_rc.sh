#!/bin/bash

cd `dirname $0`
source ../cloudrc

vm_ID=$1
action=$2

vm_xml=$xml_dir/$vm_ID.xml
[ ! -f $vm_xml ] && exit 0

cpu=$(xmllint --xpath 'string(/domain/vcpu)' $vm_xml)
let memory=$(xmllint --xpath 'string(/domain/memory)' $vm_xml)/1024
vm_img=${image_dir}/$vm_ID.qcow2
let disk=$(qemu-img info $vm_img | grep 'virtual size' | cut -d' ' -f4 | tr -d '(')/1024/1024/1024
network=0

if [ "$action" = "release" ]; then
    let cpu=-$cpu
    let memory=-$memory
    let disk=-$disk
    let network=-$network
fi

sql_exec "UPDATE cpu set used = used + $cpu"
sql_exec "UPDATE memory set used = used + $memory"
sql_exec "UPDATE disk set used = used + $disk"
sql_exec "UPDATE network set used = used + $network"
