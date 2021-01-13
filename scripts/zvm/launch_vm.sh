#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 6 ] && die "$0 <vm_ID> <image> <name> <cpu> <memory> <disk_size> <swap_size> <ephemeral_size>"

ID=$1
vm_ID=$(printf $guest_userid_template $1)
img_name=$2
vm_name=$3
vm_cpu=$4
vm_mem=$5
disk_size=$6
swap_size=$7
ephemeral_size=$8
vm_stat=error
vm_vnc=""

md=$(cat)
metadata=$(echo $md | base64 -d)

vswitch=$(jq .zvm[0].vswitch <<< $metadata | tr -d '"')
os_version=$(jq .zvm[0].osVersion <<< $metadata | tr -d '"')
disk_type=$(jq .zvm[0].diskType <<< $metadata | tr -d '"')
virt_type=$(jq .virt_type <<< $metadata | tr -d '"')
dns=$(jq .dns <<< $metadata | tr -d '"')
if [ -z "$dns" ]; then
    dns=$dns_server
fi
if [ -z "$dns" ]; then
    dns='8.8.8.8'
fi

# import image if it doesn't exist
img_name_json="\"$img_name\""
imageExists=$(curl $zvm_service/images | jq ".output |.[] | select(.imagename == $img_name_json) ")
if [ ! -n "$imageExists" ]; then
    echo "Import image $img_name"
    rc=0

    if [ ! -f "$image_cache/$img_name" ]; then
        wget -q $image_repo/$img_name -O $image_cache/$img_name
    fi
    if [ ! -f "$image_cache/$img_name" ]; then
        echo "Sync image $img_name failed!"
        echo "|:-COMMAND-:| `basename $0` '$ID' '$vm_stat' '$SCI_CLIENT_ID' 'Sync image $img_name failed!'"
        exit -1
    fi 

    if [ $os_version = "rhcos4" ]; then
        rc=`curl -s $zvm_service/images -X POST -d '{"image": {"url": "'"$image_repo/$img_name"'", "image_meta": {"os_version": "'"$os_version"'", "disk_type": "'"$disk_type"'"}, "image_name": "'"$img_name"'"}}' | jq .rc`        
    else
        rc=`curl -s $zvm_service/images -X POST -d '{"image": {"url": "'"$image_repo/$img_name"'", "image_meta": {"os_version": "'"$os_version"'"}, "image_name": "'"$img_name"'"}}' | jq .rc`
    fi

    rm -f $image_cache/$img_name

    if [ $rc -ne 0 ]; then
        echo "Import image $img_name failed!"
        echo "|:-COMMAND-:| `basename $0` '$ID' '$vm_stat' '$SCI_CLIENT_ID' 'Import image $img_name failed!'"
        exit -1
    fi
fi

./build_meta.sh "$vm_ID" "$vm_name" <<< $md >/dev/null 2>&1

# get network information from metadata
ip_address=$(jq .networks[0].ip_address <<< $metadata | tr -d '"')
netmask=$(jq .networks[0].netmask <<< $metadata | tr -d '"')
gateway=$(jq .networks[0].routes[0].gateway <<< $metadata | tr -d '"')
network=$(ipcalc -n $ip_address $netmask | tr -d NETWORK=)
prefix=$(ipcalc -p $ip_address $netmask | tr -d PREFIX=)
cidr="$network/$prefix"

mkdir -p /tmp/cloudland/pending
touch /tmp/cloudland/pending/$vm_ID

# create guest
if [ $vm_mem -lt 9999 ]; then
    rc=`curl -s $zvm_service/guests -X POST -d '{"guest":{"userid":"'"$vm_ID"'", "vcpus":'$vm_cpu', "max_cpu":'$vm_cpu', "memory":'$vm_mem', "max_mem":"'"${vm_mem}"'M", "ipl_from":"100"}}' | jq .rc`
else
    let vm_mem=$vm_mem/1024
    rc=`curl -s $zvm_service/guests -X POST -d '{"guest":{"userid":"'"$vm_ID"'", "vcpus":'$vm_cpu', "max_cpu":'$vm_cpu', "memory":'$vm_mem', "max_mem":"'"${vm_mem}"'G", "ipl_from":"100"}}' | jq .rc`
fi
if [ $rc -ne 0 ]; then
    echo "Create $vm_ID failed!"
    echo "|:-COMMAND-:| `basename $0` '$ID' '$vm_stat' '$SCI_CLIENT_ID' 'Create $vm_ID failed!'"
    rm -f /tmp/cloudland/pending/$vm_ID
    exit -1
fi

# add disk: 
# note: frontend sends disk size counting in G now. 
rc=$(curl -s $zvm_service/guests/$vm_ID/disks -X POST -d '{"disk_info":{"disk_list":[{"size":"'"$disk_size"'G", "is_boot_disk":"True"}]}}' | jq .rc)
if [ $rc -ne 0 ]; then
    # remove user ?
    echo "$vm_ID: Add disk failed!"
    echo "|:-COMMAND-:| `basename $0` '$ID' '$vm_stat' '$SCI_CLIENT_ID' '$vm_ID: Add disk failed!'"
    rm -f /tmp/cloudland/pending/$vm_ID
    exit -1
fi

# deploy
vm_meta="${cache_dir}/meta/${vm_ID}/cfgdrive.iso"
rc=$(curl -s $zvm_service/guests/$vm_ID/action -X POST -d '{"action":"deploy", "image":"'"$img_name"'", "transportfiles":"'"$vm_meta"'"}' | jq .rc)
if [ $rc -ne 0 ]; then
    # remove disk and user ?
    echo "$vm_ID: Deploy image failed!"
    echo "|:-COMMAND-:| `basename $0` '$ID' '$vm_stat' '$SCI_CLIENT_ID' '$vm_ID: Deploy image failed!'"
    rm -f /tmp/cloudland/pending/$vm_ID
    exit -1
fi

# create network interface
rc=$(curl -s $zvm_service/guests/$vm_ID/interface -X POST -d '{"interface":{"os_version":"'"$os_version"'", "guest_networks":[{"ip_addr":"'"$ip_address"'", "dns_addr":["'"$dns"'"], "gateway_addr":"'"$gateway"'", "cidr":"'"$cidr"'"}]}}'  | jq .rc)
if [ $rc -ne 0 ]; then
    # remove disk and user ?
    echo "$vm_ID: Create network failed!"
    echo "|:-COMMAND-:| `basename $0` '$ID' '$vm_stat' '$SCI_CLIENT_ID' '$vm_ID: Create network failed!'"
    rm -f /tmp/cloudland/pending/$vm_ID
    exit -1
fi

# couple to vswitch
rc=$(curl -s $zvm_service/guests/$vm_ID/nic/1000 -X PUT -d '{"info":{"couple": "True", "active": "False", "vswitch": "'"$vswitch"'"}}' | jq .rc)
if [ $rc -ne 0 ]; then
    # remove disk and user ?
    echo "$vm_ID: Couple to vswitch failed!"
    echo "|:-COMMAND-:| `basename $0` '$ID' '$vm_stat' '$SCI_CLIENT_ID' '$vm_ID: Couple to vswitch failed!'"
    rm -f /tmp/cloudland/pending/$vm_ID
    exit -1
fi

# grant user to vswitch
rc=$(curl -s $zvm_service/vswitches/$vswitch -X PUT -d '{"vswitch":{"grant_userid": "'"$vm_ID"'"}}' | jq .rc)
if [ $rc -ne 0 ]; then
    # remove disk and user ?
    echo "$vm_ID: vswitch grant failed!"
    echo "|:-COMMAND-:| `basename $0` '$ID' '$vm_stat' '$SCI_CLIENT_ID' '$vm_ID: vswitch grant failed!'"
    rm -f /tmp/cloudland/pending/$vm_ID
    exit -1
fi

# start VM
rc=$(curl -s $zvm_service/guests/$vm_ID/action -X POST -d '{"action":"start"}' | jq .rc)
if [ $rc -ne 0 ]; then
    # remove disk and user ?
    echo "$vm_ID: Start VM failed!"
    echo "|:-COMMAND-:| `basename $0` '$ID' '$vm_stat' '$SCI_CLIENT_ID' '$vm_ID: Start VM failed!'"
    rm -f /tmp/cloudland/pending/$vm_ID
    exit -1
fi

rm -f /tmp/cloudland/pending/$vm_ID

vm_stat=running
echo "|:-COMMAND-:| $(basename $0) '$ID' '$vm_stat' '$SCI_CLIENT_ID' 'unknown'"
