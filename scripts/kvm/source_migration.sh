#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 3 ] && die "$0 <vm_ID> <target_hyper> <migration_type>"

ID=$1
vm_ID=inst-$1
target_hyper=$2
migration_type=$3
state=error

if [ "$migration_type" = "warm" ]; then
    virsh migrate --live $vm_ID qemu+ssh://$target_hyper/system
    [ $? -eq 0 ] && state="vm_migrated"
else
    virsh shutdown $vm_ID
    for i in {1..60}; do
        vm_state=$(virsh dominfo $vm_ID | grep State | cut -d: -f2- | xargs | sed 's/shut off/shut_off/g')
        [ "$vm_state" = "shut_off" ] && break
        sleep 0.5
    done
    if [ "$vm_state" != "shut_off" ]; then
        virsh destroy $vm_ID
    fi
    state="source_prepared"
    echo "|:-COMMAND-:| $(basename $0) '$ID' '$SCI_CLIENT_ID' '$state'"
fi
