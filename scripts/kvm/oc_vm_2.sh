#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 1 ] && die "$0 <vm_ID>"

ID=$1
vm_ID=inst-$1
vm_xml=$xml_dir/$vm_ID/${vm_ID}.xml

count=0
while [ $count -le 100 ]; do
    sleep 5
    virsh list | grep $vm_ID
    [ $? -ne 0 ] && break
    let count=$count+1
done
if [ $? -eq 0 ]; then
    virsh dumpxml $vm_ID 2>/dev/null > ${vm_xml}.dump
    mv -f ${vm_xml}.dump $vm_xml
    virsh undefine $vm_ID
    sed -i "/initrd/d;/kernel/d;/cmdline/d;s/<on_reboot>destroy/<on_reboot>restart/;s/<on_crash>destroy/<on_crash>restart/" ${vm_xml}
    virsh define $vm_xml
    virsh start $vm_ID
    virsh autostart $vm_ID
fi
