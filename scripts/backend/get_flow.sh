#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && die "$0 <vm_id> <vlan>"

vm_id=$1
vlan=$2
# Currently there will be only one interface per VM
if_name=$3

br_name=br$vlan
ovs-vsctl br-exists $br_name
if [ $? -ne 0 ]; then
    echo "|:-COMMAND-:| `basename $0` 'switch $vlan is not existed'"
    exit -1
fi

if [ -z "$if_name" ]; then
    vm_xml=$(virsh dumpxml $vm_id)
    if [ $? -ne 0 ]; then
        echo "|:-COMMAND-:| `basename $0` 'VM $vm_id is not found'"
        exit 0
    fi

    count=$(echo $vm_xml | xmllint --xpath 'count(/domain/devices/interface)' -)
    for (( i=1; i <= $count; i++ )); do
        vif_dev=$(echo $vm_xml | xmllint --xpath "string(/domain/devices/interface[$i]/target/@dev)" -)
        flows="$flows $(ovs-ofctl dump-flows $br_name --names | grep $vif_dev)"
    done
else
    flows="$(ovs-ofctl dump-flows $br_name --names | grep $if_name)"
fi

flows="$flows $(ovs-ofctl dump-flows $br_name --names | grep ${br_name}vx4789)"

echo "|:-COMMAND-:| `basename $0` '`echo $flows`'"
