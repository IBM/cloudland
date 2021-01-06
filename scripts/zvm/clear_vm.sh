#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 1 ] && die "$0 <vm_ID>"

vm_ID=$(printf $guest_userid_template $1)

rc=$(curl -s $zvm_service/guests/$vm_ID/action -X POST -d '{"action":"stop"}' | jq .rc)

state=$(curl -s $zvm_service/guests/$vm_ID/power_state | jq .output | tr -d '"')
if [ $state != "off"  ]; then
    echo "$vm_ID stop failed."
else
    rc=$(curl -s $zvm_service/guests/$vm_ID -X DELETE | jq .rc)
    if [ $rc -ne 0 ]; then
        echo "Delete $vm_ID failed."
    fi

    rm -rf /tmp/cloudland/${vm_ID}
    rm -rf ${cache_dir}/meta/${vm_ID}
    rm -rf ${xml_dir}/${vm_ID}
fi
echo "|:-COMMAND-:| $(basename $0) '$1'"
