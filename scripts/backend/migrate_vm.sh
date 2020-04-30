#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && die "$0 <vm_ID> <hyper>"

ID=$1
vm_ID=inst-$ID
hyper=$2
xml_dir=$xml_dir/$vm_ID
[ -d "$xml_dir" ] && copy_target $xml_dir $hyper
inst_disk=$image_dir/${vm_ID}.disk
[ -f "$inst_disk" ] && copy_target $inst_disk $hyper
ephemeral=$image_dir/${vm_ID}.ephemeral
[ -f "$ephemeral" ] && copy_target $ephemeral $hyper
metaiso=$cache_dir/meta/${vm_ID}.iso
[ -f "$metaiso" ] && copy_target $metaiso $hyper
action_target $hyper "sudo virsh define $xml_dir/${vm_ID}.xml && sudo virsh start $vm_ID"
state=$(action_target $hyper "sudo virsh dominfo $vm_ID" | grep State | cut -d: -f2- | xargs)
if [ "$state" = "running" ]; then
    mv $xml_dir $backup_dir
    mv $inst_disk $ephemeral $metaiso $backup_dir/$vm_ID
    ./clear_vm.sh $ID
else
    virsh undefine $vm_ID
fi
