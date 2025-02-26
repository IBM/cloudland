#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 8 ] && die "$0 <vm_ID> <image> <snapshot> <volume_id> <old_volume_uuid> <cpu> <memory> <disk_size>"

ID=$1
vm_ID=inst-$ID
img_name=$2
snapshot=$3
vol_ID=$4
old_volume_id=$5
vm_cpu=$6
vm_mem=$7
disk_size=$8

vm_xml=$xml_dir/$vm_ID/${vm_ID}.xml
mv $vm_xml $vm_xml-$(date +'%s.%N')
timeout_virsh dumpxml $vm_ID >$vm_xml
timeout_virsh undefine $vm_ID
virsh destroy $vm_ID
let fsize=$disk_size*1024*1024*1024
if [ -z "$wds_address" ]; then
    vm_img=$volume_dir/$vm_ID.disk
    if [ ! -f "$vm_img" ]; then
        vm_img=$image_dir/$vm_ID.disk
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
    fi
else
    get_wds_token
    image=$(basename $img_name .raw)
    old_vhost_name=$(basename $(ls /var/run/wds/instance-$ID-volume-$vol_ID-*))
    vhost_id=$(wds_curl GET "api/v2/sync/block/vhost?name=$old_vhost_name" | jq -r '.vhosts[0].id')
    uss_id=$(get_uss_gateway)
    wds_curl PUT "api/v2/sync/block/vhost/unbind_uss" "{\"vhost_id\": \"$vhost_id\", \"uss_gw_id\": \"$uss_id\", \"is_snapshot\": false}"
    wds_curl DELETE "api/v2/sync/block/vhost/$vhost_id"
    wds_curl DELETE "api/v2/sync/block/volumes/$old_volume_id?force=true"

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

    for i in {1..10}; do
        vhost_name=instance-$ID-volume-$vol_ID-$RANDOM
	[ "$vhost_name" != "$old_vhost_name" ] && break
    done
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
    vhost_ret=$(wds_curl POST "api/v2/sync/block/vhost" "{\"name\": \"$vhost_name\"}")
    vhost_id=$(echo $vhost_ret | jq -r .id)
    uss_ret=$(wds_curl PUT "api/v2/sync/block/vhost/bind_uss" "{\"vhost_id\": \"$vhost_id\", \"uss_gw_id\": \"$uss_id\", \"lun_id\": \"$volume_id\", \"is_snapshot\": false}")
    ret_code=$(echo $uss_ret | jq -r .ret_code)
    if [ "$ret_code" != "0" ]; then
        echo "|:-COMMAND-:| `basename $0` '$ID' '$state' '$SCI_CLIENT_ID' 'failed to create wds vhost for boot volume, $vhost_ret, $uss_ret!'"
        exit -1
    fi
    echo "|:-COMMAND-:| create_volume_wds_vhost '$vol_ID' 'attached' 'wds_vhost://$wds_pool_id/$volume_id'"
fi

[ -z "$vm_mem" ] && vm_mem='1024m'
[ -z "$vm_cpu" ] && vm_cpu=1
let vm_mem=${vm_mem%[m|M]}*1024
sed_cmd="s#>.*</memory>#>$vm_mem</memory>#g; s#>.*</currentMemory>#>$vm_mem</currentMemory>#g; s#>.*</vcpu>#>$vm_cpu</vcpu>#g"
if [ -n "$wds_address" ]; then
  sed_cmd="$sed_cmd; s#$old_vhost_name#$vhost_name#g"
fi
sed -i "$sed_cmd" $vm_xml
timeout_virsh define $vm_xml
timeout_virsh autostart $vm_ID
timeout_virsh start $vm_ID
[ $? -eq 0 ] && state=running
echo "|:-COMMAND-:| launch_vm.sh '$ID' '$state' '$SCI_CLIENT_ID' 'sync'"
