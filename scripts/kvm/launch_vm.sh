#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 6 ] && die "$0 <vm_ID> <image> <snapshot> <name> <cpu> <memory> <disk_size> <volume_id>"

ID=$1
vm_ID=inst-$1
img_name=$2
snapshot=$3
vm_name=$4
vm_cpu=$5
vm_mem=$6
disk_size=$7
vol_ID=$8
state=error
vm_vnc=""

md=$(cat)
metadata=$(echo $md | base64 -d)

let fsize=$disk_size*1024*1024*1024
./build_meta.sh "$vm_ID" "$vm_name" <<< $md >/dev/null 2>&1
vm_meta=$cache_dir/meta/$vm_ID.iso
template=$template_dir/template.xml

if [ -z "$wds_address" ]; then
    vm_img=$volume_dir/$vm_ID.disk
    is_vol="true"
    if [ ! -f "$vm_img" ]; then
        vm_img=$image_dir/$vm_ID.disk
        is_vol="false"
        if [ ! -s "$image_cache/$img_name" ]; then
            echo "Image is not available!"
            echo "|:-COMMAND-:| `basename $0` '$ID' '$state' '$SCI_CLIENT_ID' 'image $img_name not available!'"
            exit -1
        fi
        format=$(qemu-img info $image_cache/$img_name | grep 'file format' | cut -d' ' -f3)
        cmd="qemu-img convert -f $format -O qcow2 $image_cache/$img_name $vm_img"
        result=$(eval "$cmd")
        vsize=$(qemu-img info $vm_img | grep 'virtual size:' | cut -d' ' -f5 | tr -d '(')
        if [ "$vsize" -gt "$fsize" ]; then
            echo "|:-COMMAND-:| `basename $0` '$ID' '$vm_stat' '$SCI_CLIENT_ID' 'flavor is smaller than image size'"
            exit -1
        fi
        qemu-img resize -q $vm_img "${disk_size}G" &> /dev/null
        echo "|:-COMMAND-:| create_volume_local.sh '$vol_ID' 'volume-${vol_ID}.disk' 'attached'"
    fi
else
    image=$(basename $img_name .raw)
    vhost_name=instance-$ID-volume-$vol_ID
    snapshot_name=${image}-${snapshot}
    snapshot_id=$(wds_curl GET "api/v2/sync/block/snaps?name=$snapshot_name" | jq -r '.snaps[0].id')
    if [ -z "$snapshot_id" -o "$snapshot_id" = null ]; then
	image_volume_id=$(wds_curl GET "api/v2/sync/block/volumes?name=$image" | jq -r '.volumes[0].id')
	wds_curl POST "api/v2/sync/block/snaps" "{\"name\": \"$snapshot_name\", \"description\": \"$snapshot_name\", \"volume_id\": \"$image_volume_id\"}"
        snapshot_id=$(wds_curl GET "api/v2/sync/block/snaps?name=$snapshot_name" | jq -r '.snaps[0].id')
        if [ -z "$snapshot_id" -o "$snapshot_id" = null ]; then
            echo "|:-COMMAND-:| `basename $0` '$ID' '$state' '$SCI_CLIENT_ID' 'failed to create image snapshot'"
            exit -1
        fi
        wds_curl DELETE "api/v2/sync/block/snaps/$image-$(($snapshot-1))?force=false"
    fi
    volume_id=$(wds_curl POST "api/v2/sync/block/snaps/$snapshot_id/clone" "{\"name\": \"$vhost_name\"}" | jq -r .id)
    rest_code=$(wds_curl PUT "api/v2/sync/block/volumes/$volume_id/expand" "{\"size\": $fsize}" | jq -r .ret_code)
    if [ "$rest_code" != "0" ]; then
        echo "|:-COMMAND-:| `basename $0` '$ID' '$state' '$SCI_CLIENT_ID' 'failed to create boot volume!'"
        exit -1
    fi
    uss_id=$(get_uss_gateway)
    vhost_id=$(wds_curl POST "api/v2/sync/block/vhost" "{\"name\": \"$vhost_name\"}" | jq -r .id)
    ret_code=$(wds_curl PUT "api/v2/sync/block/vhost/bind_uss" "{\"vhost_id\": \"$vhost_id\", \"uss_gw_id\": \"$uss_id\", \"lun_id\": \"$volume_id\", \"is_snapshot\": false}" | jq -r .ret_code)
    if [ "$rest_code" != "0" ]; then
        echo "|:-COMMAND-:| `basename $0` '$ID' '$state' '$SCI_CLIENT_ID' 'failed to create wds vhost for boot volume!'"
        exit -1
    fi
    echo "|:-COMMAND-:| create_volume_wds_vhost '$vol_ID' 'attached' 'wds_vhost://$wds_pool_id/$volume_id'"
    ux_sock=/var/run/wds/$vhost_name
    template=$template_dir/wds_template.xml
fi

[ -z "$vm_mem" ] && vm_mem='1024m'
[ -z "$vm_cpu" ] && vm_cpu=1
let vm_mem=${vm_mem%[m|M]}*1024
mkdir -p $xml_dir/$vm_ID
vm_xml=$xml_dir/$vm_ID/${vm_ID}.xml
cp $template $vm_xml
sed -i "s/VM_ID/$vm_ID/g; s/VM_MEM/$vm_mem/g; s/VM_CPU/$vm_cpu/g; s#VM_IMG#$vm_img#g; s#VM_UNIX_SOCK#$ux_sock#g; s#VM_META#$vm_meta#g;" $vm_xml
virsh define $vm_xml
virsh autostart $vm_ID
jq .vlans <<< $metadata | ./sync_nic_info.sh "$ID"
virsh start $vm_ID
[ $? -eq 0 ] && state=running
echo "|:-COMMAND-:| $(basename $0) '$ID' '$state' '$SCI_CLIENT_ID' 'init'"
