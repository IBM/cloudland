#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <img_ID> <img_Prefix> <vm_ID> <boot_volume>" && exit -1

img_ID=$1
prefix=$2
vm_ID=inst-$3
boot_volume=$4
image_name=image-$img_ID-$prefix
state=error

# its better to let user shutdown the vm before capturing the image
# timeout_virsh suspend $vm_ID
if [ -z "$wds_address" ]; then
    # capture the image from the running instance locally
    image=${image_dir}/image-$vm_ID.qcow2
    inst_img=$cache_dir/instance/${vm_ID}.disk

    format=$(qemu-img info $inst_img | grep 'file format' | cut -d' ' -f3)
    qemu-img convert -f $format -O qcow2 $inst_img $image
    [ -s "$image" ] && state=available
    sync_target /opt/cloudland/cache/image/
else
    # clone the image from the boot volume on the remote storage WDS
    if [ -z "$boot_volume" ]; then
        echo "|:-COMMAND-:| capture_image.sh '$img_ID' 'error' 'qcow2' 'boot_volume is not specified'"
        exit -1
    fi
    get_wds_token
    # use max speed to clone the boot volume
    clone_ret=$(wds_curl PUT "api/v2/sync/block/volumes/$boot_volume/copy_clone" "{\"name\":\"$image_name\", \"speed\": 32, \"phy_pool_id\": \"$wds_pool_id\"}")
    read -d'\n' -r task_id ret_code message < <(jq -r ".task_id .ret_code .message" <<< $clone_ret)
    [ "$ret_code" != "0" ] && echo "|:-COMMAND-:| capture_image.sh '$img_ID' 'error' 'qcow2' 'failed to clone the boot volume: $message'" && exit -1
    state=cloning
    for i in {1..100}; do
        st=$(wds_curl GET "api/v2/sync/block/volumes/tasks/$task_id" | jq -r .task.state)
	    [ "$st" = "TASK_COMPLETE" ] && state=uploaded && break
	    [ "$st" = "TASK_FAILED" ] && state=failed && break
	sleep 5
    done

    volume_id=$(wds_curl GET "api/v2/sync/block/volumes?name=$image_name" | jq -r '.volumes[0].id')
    [ -n "$volume_id" ] && state=available
fi
# timeout_virsh resume $vm_ID
echo "|:-COMMAND-:| capture_image.sh '$img_ID' '$state' 'qcow2' 'success'"
