#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && die "$0 <vm_ID>"

vm_ID=$1
vm_xml=$xml_dir/$vm_ID.xml
[ -f $vm_xml ] || die "No such vm definition"
virsh desc $vm_ID &> /dev/null
if [ $? != 0 ]; then
    virsh create $vm_xml
    virsh dumpxml --security-info $vm_ID 2>/dev/null | sed "s/autoport='yes'/autoport='no'/g" > $vm_xml.dump && mv -f $vm_xml.dump $vm_xml
    [ $? -eq 0 ] || die "failed to create vm"
fi
vm_stat=running
echo "|:-COMMAND-:| /opt/cloudland/scripts/frontback/`basename $0` '$vm_ID' '$vm_stat'"
