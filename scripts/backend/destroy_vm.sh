#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && die "$0 <vm_ID> <force>"

vm_ID=$1
force=`echo ${2:0:1} | tr [T] t`
shutdown="shutdown"
[ "$force" = "t" ] && shutdown="destroy" && virsh undefine $vm_ID
virsh desc $vm_ID &> /dev/null
if [ $? = 0 ]; then
    virsh $shutdown $vm_ID
fi
status=stopped
echo "|:-COMMAND-:| $(basename $0) $vm_ID $status"
