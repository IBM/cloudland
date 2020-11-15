#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 6 ] && die "$0 <vm_ID> <image> <name> <cpu> <memory> <disk_size> <swap_size> <ephemeral_size>"

ID=$1
vm_ID=inst-$1
img_name=$2
vm_name=$3
vm_cpu=$4
vm_mem=$5
disk_size=$6
swap_size=$7
ephemeral_size=$8
vm_stat=error
vm_vnc=""

metadata=$(base64 -d)
./build_meta.sh "$vm_ID" "$vm_name" <<< $metadata >/dev/null 2>&1
vm_img=$volume_dir/$vm_ID.disk
is_vol="true"
if [ ! -f "$vm_img" ]; then
    vm_img=$image_dir/$vm_ID.disk
    vm_meta=$cache_dir/meta/$vm_ID.iso
    is_vol="false"
    if [ ! -f "$image_cache/$img_name" ]; then
        wget -q $image_repo/$img_name -O $image_cache/$img_name
    fi
    if [ ! -f "$image_cache/$img_name" ]; then
        echo "Image $img_name downlaod failed!"
        echo "|:-COMMAND-:| `basename $0` '$ID' '$vm_stat' '$SCI_CLIENT_ID' 'image $img_name downlaod failed!'"
        exit -1
    fi
    format=$(qemu-img info $image_cache/$img_name | grep 'file format' | cut -d' ' -f3)
    cmd="qemu-img convert -f $format -O qcow2 $image_cache/$img_name $vm_img"
    result=$(eval "$cmd")
    sidecar span log $span "Internal: $cmd, result: $result"
    vsize=$(qemu-img info $vm_img | grep 'virtual size:' | cut -d' ' -f4 | tr -d '(')
    let fsize=$disk_size*1024*1024*1024
    if [ "$vsize" -gt "$fsize" ]; then
        echo "|:-COMMAND-:| `basename $0` '$ID' '$vm_stat' '$SCI_CLIENT_ID' 'flavor is smaller than image size'"
        exit -1
    fi
    qemu-img resize -q $vm_img "${disk_size}G" &> /dev/null
fi

hyper_ip=$(ifconfig $vxlan_interface | grep 'inet addr:' | cut -d: -f2 | cut -d' ' -f1)
[ -z "$vm_mem" ] && vm_mem='1024m'
[ -z "$vm_cpu" ] && vm_cpu=1
let vm_mem=${vm_mem%[m|M]}*1024
mkdir -p $xml_dir/$vm_ID
vm_xml=$xml_dir/$vm_ID/${vm_ID}.xml
template=$template_dir/template.xml
[ $(uname -m) = s390x ] && template=$template_dir/linuxone.xml
cp $template $vm_xml
sed -i "s/VM_ID/$vm_ID/g; s/VM_MEM/$vm_mem/g; s/VM_CPU/$vm_cpu/g; s#VM_IMG#$vm_img#g; s#VM_META#$vm_meta#g;" $vm_xml
state=error
virsh define $vm_xml
virsh autostart $vm_ID
if [ "$ephemeral_size" -gt 0 ]; then
    ephemeral=$image_dir/${vm_ID}.ephemeral
    qemu-img create $ephemeral ${ephemeral_size}G
    virsh attach-disk $vm_ID $ephemeral vdb --config
fi
vlans=$(jq .vlans <<< $metadata)
nvlan=$(jq length <<< $vlans)
i=0
while [ $i -lt $nvlan ]; do
    vlan=$(jq -r .[$i].vlan <<< $vlans)
    ip=$(jq -r .[$i].ip_address <<< $vlans)
    mac=$(jq -r .[$i].mac_address <<< $vlans)
    jq .security <<< $metadata | ./attach_nic.sh $ID $vlan $ip $mac 
    let i=$i+1
done
virsh start $vm_ID
[ $? -eq 0 ] && state=running && ./replace_vnc_passwd.sh $ID
echo "|:-COMMAND-:| $(basename $0) '$ID' '$state' '$SCI_CLIENT_ID' 'unknown'"
