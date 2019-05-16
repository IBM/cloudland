#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <vm_ID> <volume_ID>" && exit -1

vm_ID=inst-$1
vol_ID=$2
vol_xml=$xml_dir/$vm_ID/disk-${vol_ID}.xml
virsh detach-device $vm_ID $vol_xml --config --persistent
if [ $? -eq 0 ]; then
    echo "|:-COMMAND-:| $(basename $0) '$1' '$vol_ID'"
else
    echo "|:-COMMAND-:| $(basename $0) '' '$vol_ID'"
fi
vm_xml=$xml_dir/$vm_ID/$vm_ID.xml
virsh dumpxml --security-info $vm_ID 2>/dev/null | sed "s/autoport='yes'/autoport='no'/g" > $vm_xml.dump && mv -f $vm_xml.dump $vm_xml
