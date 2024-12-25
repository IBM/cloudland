#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && die "$0 <vm_ID> <action>"

vm_ID=inst-$1
action=$2
if [ "$action" = "restart" ]; then
    virsh reboot $vm_ID
elif [ "$action" = "start" ]; then
    virsh start $vm_ID
elif [ "$action" = "stop" ]; then
    virsh shutdown $vm_ID
elif [ "$action" = "hard_stop" ]; then
    virsh destroy $vm_ID
elif [ "$action" = "restart" ]; then
    virsh reboot $vm_ID
elif [ "$action" = "hard_restart" ]; then
    virsh destroy $vm_ID
    sleep 0.5
    virsh start $vm_ID
elif [ "$action" = "pause" ]; then
    virsh suspend $vm_ID
elif [ "$action" = "resume" ]; then
    virsh resume $vm_ID
else
    die "Invalid action: $action"
fi
sleep 1
state=$(virsh dominfo $vm_ID | grep State | cut -d: -f2- | xargs | sed 's/shut off/shut_off/g')
echo "|:-COMMAND-:| $(basename $0) '$1' '$state'"
