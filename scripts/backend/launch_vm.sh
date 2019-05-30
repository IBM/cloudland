#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 6 ] && die "$0 <vm_ID> <image> <name> <cpu> <memory> <disk_inc> [userdata] [pubkey]"

vm_ID=inst-$1
img_name=$2
vm_name=$3
vm_cpu=$4
vm_mem=$5
disk_inc=$6
userdata=$7
pubkey=$8
vm_stat=error
vm_vnc=""

metadata=$(cat)
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
        echo "|:-COMMAND-:| `basename $0` '$vm_ID' '$vm_stat' '$SCI_CLIENT_ID'"
        exit -1
    fi
    cmd="qemu-img convert -f qcow2 -O raw $image_cache/$img_name $vm_img"
    result=$(eval "$cmd")
    sidecar span log $span "Internal: $cmd, result: $result"
    qemu-img resize -q $vm_img "${disk_inc}G" &> /dev/null
fi

vsize=`qemu-img info $vm_img | grep 'virtual size:' | cut -d' ' -f3`
hyper_ip=$(ifconfig $vxlan_interface | grep 'inet addr:' | cut -d: -f2 | cut -d' ' -f1)
[ -z "$vm_mem" ] && vm_mem='1024m'
[ -z "$vm_cpu" ] && vm_cpu=1
let vm_mem=${vm_mem%[m|M]}*1024
vnc_pass=`date | sum | cut -d' ' -f1`
mkdir -p $xml_dir/$vm_ID
vm_xml=$xml_dir/$vm_ID/$vm_ID.xml
template=$template_dir/template.xml
cp $template $vm_xml
sed -i "s/VM_ID/$vm_ID/g; s/VM_MEM/$vm_mem/g; s/VM_CPU/$vm_cpu/g; s#VM_IMG#$vm_img#g; s#VM_META#$vm_meta#g; s/VNC_PASS/$vnc_pass/g;" $vm_xml
state=error
virsh define $vm_xml
virsh autostart $vm_ID
vlans=$(jq .vlans <<< $metadata)
nvlan=$(jq length <<< $vlans)
i=0
while [ $i -lt $nvlan ]; do
    vlan=$(jq -r .[$i].vlan <<< $vlans)
    mac=$(jq -r .[$i].mac_address <<< $vlans)
    ./attach_nic.sh $vm_ID $vlan $mac
    let i=$i+1
done
virsh start $vm_ID
[ $? -eq 0 ] && state=running
virsh dumpxml --security-info $vm_ID 2>/dev/null | sed "s/autoport='yes'/autoport='no'/g" > $vm_xml.dump && mv -f $vm_xml.dump $vm_xml
vnc_port=$(xmllint --xpath 'string(/domain/devices/graphics/@port)' $vm_xml)
vm_vnc="$vnc_port:$vnc_pass"

echo "|:-COMMAND-:| $(basename $0) '$1' '$state' '$SCI_CLIENT_ID'"
