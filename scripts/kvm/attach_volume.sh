#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <vm_ID> <volume_ID> <volume_path>" && exit -1

vm_ID=inst-$1
vol_ID=$2
vol_path=$gluster_volume/$3
vol_xml=$xml_dir/$vm_ID/disk-${vol_ID}.xml
cp $template_dir/volume.xml $vol_xml
count=$(virsh dumpxml $vm_ID | grep -c "<disk type='network' device='disk'>")
let letter=98+$count
device=vd$(printf "\\$(printf '%03o' "$letter")")
sed -i "s#VOLUME_SOURCE#$vol_path#g;s#VOLUME_TARGET#$device#g;" $vol_xml
virsh attach-device $vm_ID $vol_xml --config --persistent
if [ $? -eq 0 ]; then
    echo "|:-COMMAND-:| $(basename $0) '$1' '$vol_ID' '$device'"
else
    echo "|:-COMMAND-:| $(basename $0) '' '$vol_ID' ''"
fi
vm_xml=$xml_dir/$vm_ID/$vm_ID.xml
virsh dumpxml --security-info $vm_ID 2>/dev/null | sed "s/autoport='yes'/autoport='no'/g" > $vm_xml.dump && mv -f $vm_xml.dump $vm_xml
