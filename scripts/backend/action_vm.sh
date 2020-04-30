#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && die "$0 <vm_ID> <action>"

vm_ID=inst-$1
action=$2
virsh $action $vm_ID
sleep 1
state=$(virsh dominfo $vm_ID | grep State | cut -d: -f2- | xargs | sed 's/shut off/shut_off/g')
echo "|:-COMMAND-:| $(basename $0) '$1' '$state'"
