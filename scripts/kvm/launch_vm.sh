#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 7 ] && die "$0 <vm_ID> <image> <qa_enabled> <snapshot> <name> <cpu> <memory> <disk_size> <volume_id>"

ID=$1
vm_ID=inst-$1
img_name=$2
qa_enabled=$3
snapshot=$4
vm_name=$5
vm_cpu=$6
vm_mem=$7
disk_size=$8
vol_ID=$9
state=error
vm_vnc=""

md=$(cat)
metadata=$(echo $md | base64 -d)

let fsize=$disk_size*1024*1024*1024
./build_meta.sh "$vm_ID" "$vm_name" <<< $md >/dev/null 2>&1
vm_meta=$cache_dir/meta/$vm_ID.iso
template=$template_dir/template_with_qa.xml
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
    get_wds_token
    image=$(basename $img_name .raw)
    vhost_name=instance-$ID-volume-$vol_ID-$RANDOM
    snapshot_name=${image}-${snapshot}
    snapshot_id=$(wds_curl GET "api/v2/sync/block/snaps?name=$snapshot_name" | jq -r '.snaps[0].id')
    if [ -z "$snapshot_id" -o "$snapshot_id" = null ]; then
	image_volume_id=$(wds_curl GET "api/v2/sync/block/volumes?name=$image" | jq -r '.volumes[0].id')
	snapshot_ret=$(wds_curl POST "api/v2/sync/block/snaps" "{\"name\": \"$snapshot_name\", \"description\": \"$snapshot_name\", \"volume_id\": \"$image_volume_id\"}")
        snapshot_id=$(wds_curl GET "api/v2/sync/block/snaps?name=$snapshot_name" | jq -r '.snaps[0].id')
        if [ -z "$snapshot_id" -o "$snapshot_id" = null ]; then
            echo "|:-COMMAND-:| `basename $0` '$ID' '$state' '$SCI_CLIENT_ID' 'failed to create image snapshot, $snapshot_ret'"
            exit -1
        fi
        wds_curl DELETE "api/v2/sync/block/snaps/$image-$(($snapshot-1))?force=false"
    fi
    volume_ret=$(wds_curl POST "api/v2/sync/block/snaps/$snapshot_id/clone" "{\"name\": \"$vhost_name\"}")
    volume_id=$(echo $volume_ret | jq -r .id)
    if [ -z "$volume_id" -o "$volume_id" = null ]; then
        echo "|:-COMMAND-:| `basename $0` '$ID' '$state' '$SCI_CLIENT_ID' 'failed to create boot volume based on snapshot $snapshot_name, $volume_ret!'"
        exit -1
    fi
    expand_ret=$(wds_curl PUT "api/v2/sync/block/volumes/$volume_id/expand" "{\"size\": $fsize}")
    ret_code=$(echo $expand_ret | jq -r .ret_code)
    if [ "$ret_code" != "0" ]; then
        echo "|:-COMMAND-:| `basename $0` '$ID' '$state' '$SCI_CLIENT_ID' 'failed to expand boot volume to size $fsize, $expand_ret'"
        exit -1
    fi
    uss_id=$(get_uss_gateway)
    vhost_ret=$(wds_curl POST "api/v2/sync/block/vhost" "{\"name\": \"$vhost_name\"}")
    vhost_id=$(echo $vhost_ret | jq -r .id)
    uss_ret=$(wds_curl PUT "api/v2/sync/block/vhost/bind_uss" "{\"vhost_id\": \"$vhost_id\", \"uss_gw_id\": \"$uss_id\", \"lun_id\": \"$volume_id\", \"is_snapshot\": false}")
    ret_code=$(echo $uss_ret | jq -r .ret_code)
    if [ "$ret_code" != "0" ]; then
        echo "|:-COMMAND-:| `basename $0` '$ID' '$state' '$SCI_CLIENT_ID' 'failed to create wds vhost for boot volume, $vhost_ret, $uss_ret!'"
        exit -1
    fi
    echo "|:-COMMAND-:| create_volume_wds_vhost '$vol_ID' 'attached' 'wds_vhost://$wds_pool_id/$volume_id'"
    ux_sock=/var/run/wds/$vhost_name
    template=$template_dir/wds_template_with_qa.xml
fi

[ -z "$vm_mem" ] && vm_mem='1024m'
[ -z "$vm_cpu" ] && vm_cpu=1
let vm_mem=${vm_mem%[m|M]}*1024
mkdir -p $xml_dir/$vm_ID
vm_QA="$qemu_agent_dir/$vm_ID.agent"
vm_xml=$xml_dir/$vm_ID/${vm_ID}.xml
cp $template $vm_xml
sed -i "s/VM_ID/$vm_ID/g; s/VM_MEM/$vm_mem/g; s/VM_CPU/$vm_cpu/g; s#VM_IMG#$vm_img#g; s#VM_UNIX_SOCK#$ux_sock#g; s#VM_META#$vm_meta#g; s#VM_AGENT#$vm_QA#g" $vm_xml
timeout_virsh define $vm_xml
timeout_virsh autostart $vm_ID
jq .vlans <<< $metadata | ./sync_nic_info.sh "$ID" "$vm_name"
timeout_virsh start $vm_ID
[ $? -eq 0 ] && state=running
echo "|:-COMMAND-:| $(basename $0) '$ID' '$state' '$SCI_CLIENT_ID' 'init'"

# check if the vm is windows and whether to change the rdp port
os_code=$(jq -r '.os_code' <<< $metadata)
if [ "$os_code" = "windows" ]; then
    rdp_port=$(jq -r '.login_port' <<< $metadata)
    if [ -n "$rdp_port" ] && [ "${rdp_port}" != "3389" ]; then
        # run the script to change the rdp port in background
        async_exec ./async_job/win_rdp_port.sh $vm_ID $rdp_port
    fi
fi
