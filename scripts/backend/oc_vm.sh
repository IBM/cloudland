#!/bin/bash -xv

cd $(dirname $0)
source ../cloudrc

[ $# -lt 4 ] && die "$0 <vm_ID> <cpu> <memory> <disk_size>"

vm_ID=inst-$1
vm_cpu=$2
vm_mem=$3
disk_size=$4
vm_stat=error
vm_vnc=""

vm_disk=$image_dir/$vm_ID.disk
rm -f $vm_disk
qemu-img create $vm_disk -f qcow2 "${disk_size}G"

metadata=$(cat)
[ -z "$vm_mem" ] && vm_mem='1024m'
[ -z "$vm_cpu" ] && vm_cpu=1
let vm_mem=${vm_mem%[m|M]}*1024
vnc_pass=`date | sum | cut -d' ' -f1`
mkdir -p $xml_dir/$vm_ID
vm_xml=$xml_dir/$vm_ID/$vm_ID.xml
template=$template_dir/openshift.xml
cp $template $vm_xml
sed -i "s/VM_ID/$vm_ID/g; s/VM_MEM/$vm_mem/g; s/VM_CPU/$vm_cpu/g; s#VM_IMG#$vm_disk#g; s/VNC_PASS/$vnc_pass/g;" $vm_xml
state=error
virsh create $vm_xml --paused
vlans=$(jq .vlans <<< $metadata)
nvlan=$(jq length <<< $vlans)
i=0
while [ $i -lt $nvlan ]; do
    vlan=$(jq -r .[$i].vlan <<< $vlans)
    ip=$(jq -r .[$i].ip_address <<< $vlans)
    mac=$(jq -r .[$i].mac_address <<< $vlans)
    jq .security <<< $metadata | ./attach_nic.sh $1 $vlan $ip $mac 
    let i=$i+1
done
virsh resume $vm_ID
count=0
while [ $count -le 100 ]; do
    sleep 5
    virsh list | grep $vm_ID
    [ $? -ne 0 ] && break
    let count=$count+1
done
if [ $? -eq 0 ]; then
    state=running
    virsh dumpxml --security-info $vm_ID 2>/dev/null | sed "s/autoport='yes'/autoport='no'/g" > ${vm_xml}.dump
    sed "/initrd/d;/kernel/d;/cmdline/d" ${vm_xml}.dump
    mv -f ${vm_xml}.dump $vm_xml
    vnc_port=$(xmllint --xpath 'string(/domain/devices/graphics/@port)' $vm_xml)
    vm_vnc="$vnc_port:$vnc_pass"
    virsh define $vm_xml
    virsh autostart $vm_ID
fi
echo "|:-COMMAND-:| $(basename $0) '$1' '$state' '$SCI_CLIENT_ID' 'unknown'"
