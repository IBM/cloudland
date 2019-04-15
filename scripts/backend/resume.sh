#!/bin/bash

cd `dirname $0`
source ../cloudrc

exec >/dev/null 2>&1
[ $# -lt 1 ] && die "$0 <vm_ID>"

vm_ID=$1
working_dir=/tmp/$vm_ID
mkisofs -quiet -R -V config-2 -o ${cache_dir}/meta/${vm_ID}.iso $working_dir &> /dev/null
#rm -rf $working_dir
sidecar span log $span "Internal: virsh resume $vm_ID" "Result: $result"
virsh start $vm_ID
if [ $? -eq 0 ]; then
    virsh autostart $vm_ID
    virsh dumpxml --security-info $vm_ID 2>/dev/null | sed "s/autoport='yes'/autoport='no'/g" > $vm_xml.dump && mv -f $vm_xml.dump $vm_xml
    vnc_port=$(xmllint --xpath 'string(/domain/devices/graphics/@port)' $vm_xml)
    vm_vnc="$vnc_port:$vnc_pass"
fi
