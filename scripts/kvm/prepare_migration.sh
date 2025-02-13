#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 5 ] && die "$0 <vm_ID> <name> <cpu> <memory> <disk_size>"

ID=$1
vm_ID=inst-$1
vm_name=$2
vm_cpu=$3
vm_mem=$4
disk_size=$5
state=error
vm_vnc=""

md=$(cat)
metadata=$(echo $md | base64 -d)

let fsize=$disk_size*1024*1024*1024
./build_meta.sh "$vm_ID" "$vm_name" <<< $md >/dev/null 2>&1
vm_meta=$cache_dir/meta/$vm_ID.iso
template=$template_dir/template_with_qa.xml
if [ -n "$wds_address" ]; then
    get_wds_token
    volumes=$(jq -r .volumes <<< $metadata)
    nvolume=$(jq length <<< $volumes)
    while [ $i -lt $nvlan ]; do
        read -d'\n' -r vol_ID volume_id device booting < <(jq -r ".[$i].id, .[$i].uuid, .[$i].device, .[$i].booting" <<<$volumes)
	ux_sock=/var/run/wds/$vhost_name
	[ "$booting" == "true" ] && boot_ux_sock=$ux_sock
	vpath=$(wds_curl GET "api/v2/sync/block/volumes/$volume_id/bind_status" | jq -r .path[0])
	wds_curl PUT "api/v2/failure_domain/black_list" "{\"path\": \"$vpath\"}"
	vhost_name=$(wds_curl GET "api/v2/sync/block/volumes/$vpath" | jq -r .volume_detail.name)
        vhost_id=$(wds_curl GET "api/v2/sync/block/vhost?name=$vhost_name" | jq -r '.vhosts[0].id')
	uss_ret=$(wds_curl PUT "api/v2/sync/block/vhost/bind_uss" "{\"vhost_id\": \"$vhost_id\", \"uss_gw_id\": \"$uss_id\", \"lun_id\": \"$volume_id\", \"is_snapshot\": false}")
	vol_xml=$xml_dir/$vm_ID/disk-${vol_ID}.xml
        cp $template_dir/wds_volume.xml $vol_xml
        sed -i "s#VM_UNIX_SOCK#$ux_sock#g;s#VOLUME_TARGET#$device#g;" $vol_xml
    done
    template=$template_dir/wds_template_with_qa.xml
fi

[ -z "$vm_mem" ] && vm_mem='1024m'
[ -z "$vm_cpu" ] && vm_cpu=1
let vm_mem=${vm_mem%[m|M]}*1024
mkdir -p $xml_dir/$vm_ID
vm_QA="$qemu_agent_dir/$vm_ID.agent"
vm_xml=$xml_dir/$vm_ID/${vm_ID}.xml
cp $template $vm_xml
sed -i "s/VM_ID/$vm_ID/g; s/VM_MEM/$vm_mem/g; s/VM_CPU/$vm_cpu/g; s#VM_IMG#$vm_img#g; s#VM_UNIX_SOCK#$boot_ux_sock#g; s#VM_META#$vm_meta#g; s#VM_AGENT#$vm_QA#g" $vm_xml
virsh define $vm_xml
virsh autostart $vm_ID
jq .vlans <<< $metadata | ./sync_nic_info.sh "$ID" "$vm_name"
[ $? -eq 0 ] && state=migration_prepared
echo "|:-COMMAND-:| $(basename $0) '$ID' '$state' '$SCI_CLIENT_ID'
