#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && die "$0 <vm_ID> <action>"

function wait_vm_status()
{
    vm_ID=$1
    status=$2
    # wait for 30 seconds
    for i in {1..60}; do
        state=$(virsh dominfo $vm_ID | grep State | cut -d: -f2- | xargs | sed 's/shut off/shut_off/g')
        [ "$state" = "$status" ] && break
        sleep 0.5
    done
}

vm_ID=inst-$1
action=$2
if [ "$action" = "restart" ]; then
    virsh reboot $vm_ID
    wait_vm_status $vm_ID "running"
elif [ "$action" = "start" ]; then
    virsh start $vm_ID
    wait_vm_status $vm_ID "running"
    echo "|:-COMMAND-:| launch_vm.sh '$vm_ID' 'running' '$SCI_CLIENT_ID' 'sync'"
elif [ "$action" = "stop" ]; then
    virsh shutdown $vm_ID
    wait_vm_status $vm_ID "shut_off"
elif [ "$action" = "hard_stop" ]; then
    virsh destroy $vm_ID
    wait_vm_status $vm_ID "shut_off"
elif [ "$action" = "hard_restart" ]; then
    virsh destroy $vm_ID
    wait_vm_status $vm_ID "shut_off"
    virsh start $vm_ID
elif [ "$action" = "pause" ]; then
    virsh suspend $vm_ID
    wait_vm_status $vm_ID "paused"
elif [ "$action" = "resume" ]; then
    virsh resume $vm_ID
    wait_vm_status $vm_ID "running"
    echo "|:-COMMAND-:| launch_vm.sh '$vm_ID' 'running' '$SCI_CLIENT_ID' 'sync'"
else
    die "Invalid action: $action"
fi

state=$(virsh dominfo $vm_ID | grep State | cut -d: -f2- | xargs | sed 's/shut off/shut_off/g')
echo "|:-COMMAND-:| $(basename $0) '$1' '$state'"
