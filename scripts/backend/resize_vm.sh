#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 6 ] && die "$0 <vm_ID> <cpu> <memory> <disk_size> <swap_size> <ephemeral_size>"

ID=$1
vm_ID=inst-$1
vm_cpu=$2
vm_mem=$3
disk_size=$4
swap_size=$5
ephemeral_size=$6

./action_vm.sh $ID shutdown
let vm_mem=${vm_mem%[m|M]}*1024
virsh setmaxmem $vm_ID $vm_mem --config
virsh setmem $vm_ID $vm_mem --config
virsh setvcpus $vm_ID --count $vm_cpu --config
virsh setvcpus $vm_ID --count $vm_cpu --config --maximum
virsh setvcpus $vm_ID --count $vm_cpu --config
vm_img=$image_dir/$vm_ID.disk
vsize=$(qemu-img info $vm_img | grep 'virtual size:' | cut -d' ' -f4 | tr -d '(')
let fsize=$disk_size*1024*1024*1024
[ $fsize -gt $vsize ] && qemu-img resize -q $vm_img "${disk_size}G" &> /dev/null
if [ $ephemeral_size -gt 0 ]; then
    ephemeral=$image_dir/${vm_ID}.ephemeral
    if [ ! -f "$ephemeral" ]; then
        qemu-img create $ephemeral ${ephemeral_size}G
        virsh attach-disk $vm_ID $ephemeral vdb --config
    else
        vsize=$(qemu-img info $ephemeral | grep 'virtual size:' | cut -d' ' -f4 | tr -d '(')
        let fsize=$ephemeral_size*1024*1024*1024
        [ $fsize -gt $vsize ] && qemu-img resize -q $ephemeral "${ephemeral_size}G" &> /dev/null
    fi
fi
virsh start $vm_ID
[ $? -eq 0 ] && virsh dumpxml $vm_ID > $xml_dir/$vm_ID/${vm_ID}.xml
echo "|:-COMMAND-:| inst_status.sh '$SCI_CLIENT_ID' '$ID running'"
