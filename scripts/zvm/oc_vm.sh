#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 5 ] && die "$0 <vm_ID> <cpu> <memory> <disk_size> <hostname>"

ID=$1
vm_ID=$(printf $guest_userid_template $1)
vm_cpu=$2
vm_mem=$3
disk_size=$4
hname=$5
role='worker'
[ "${5/master/}" != "$5" ] && role='master'
[ "${5/bootstrap/}" != "$5" ] && role='bootstrap'
[ -z "$role" ] && role='worker'
vm_stat=error
vm_vnc=""

# get information from metadata
metadata=$(cat)
vswitch=$(jq .zvm[0].vswitch <<< $metadata | tr -d '"')
ocp_version=$(jq .ocp[0].ocpVersion <<< $metadata | tr -d '"')
service=$(jq .ocp[0].service <<< $metadata | tr -d '"')
os_version=rhcos4
virt_type=$(jq .virt_type <<< $metadata | tr -d '"')
dns=$(jq .dns <<< $metadata | tr -d '"')
if [ -z "$dns" ]; then
    dns=$dns_server
fi
if [ -z "$dns" ]; then
    dns='8.8.8.8'
fi

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
rc=`curl -s $zvm_service/guests -X POST -d '{"guest":{"userid":"'"$vm_ID"'", "vcpus":'$vm_cpu', "max_cpu":'$vm_cpu', "memory":'$vm_mem', "max_mem":"'"${vm_mem}"'M", "ipl_from":"c"}}' | jq .rc`
if [ $rc -ne 0 ]; then
    echo "Create $vm_ID failed!"
    echo "|:-COMMAND-:| launch_vm.sh '$ID' '$vm_stat' '$SCI_CLIENT_ID' 'Create $vm_ID failed!'"
    rm -f /tmp/cloudland/pending/$vm_ID
    exit -1
fi

# add disk: 
rc=$(curl -s $zvm_service/guests/$vm_ID/disks -X POST -d '{"disk_info":{"disk_list":[{"size":"'"$disk_size"'G", "is_boot_disk":"True"}]}}' | jq .rc)
if [ $rc -ne 0 ]; then
    # remove user ?
    echo "$vm_ID: Add disk failed!"
    echo "|:-COMMAND-:| launch_vm.sh '$ID' '$vm_stat' '$SCI_CLIENT_ID' '$vm_ID: Add disk failed!'"
    rm -f /tmp/cloudland/pending/$vm_ID
    exit -1
fi

# create network interface
rc=$(curl -s $zvm_service/guests/$vm_ID/interface -X POST -d '{"interface":{"os_version":"'"$os_version"'", "guest_networks":[{"ip_addr":"'"$ip_address"'", "dns_addr":["'"$dns"'"], "gateway_addr":"'"$gateway"'", "cidr":"'"$cidr"'"}]}}'  | jq .rc)
if [ $rc -ne 0 ]; then
    # remove disk and user ?
    echo "$vm_ID: Create network failed!"
    echo "|:-COMMAND-:| launch_vm.sh '$ID' '$vm_stat' '$SCI_CLIENT_ID' '$vm_ID: Create network failed!'"
    rm -f /tmp/cloudland/pending/$vm_ID
    exit -1
fi

# couple to vswitch
rc=$(curl -s $zvm_service/guests/$vm_ID/nic/1000 -X PUT -d '{"info":{"couple": "True", "active": "False", "vswitch": "'"$vswitch"'"}}' | jq .rc)
if [ $rc -ne 0 ]; then
    # remove disk and user ?
    echo "$vm_ID: Couple to vswitch failed!"
    echo "|:-COMMAND-:| launch_vm.sh '$ID' '$vm_stat' '$SCI_CLIENT_ID' '$vm_ID: Couple to vswitch failed!'"
    rm -f /tmp/cloudland/pending/$vm_ID
    exit -1
fi

# grant user to vswitch
rc=$(curl -s $zvm_service/vswitches/$vswitch -X PUT -d '{"vswitch":{"grant_userid": "'"$vm_ID"'"}}' | jq .rc)
if [ $rc -ne 0 ]; then
    # remove disk and user ?
    echo "$vm_ID: vswitch grant failed!"
    echo "|:-COMMAND-:| launch_vm.sh '$ID' '$vm_stat' '$SCI_CLIENT_ID' '$vm_ID: vswitch grant failed!'"
    rm -f /tmp/cloudland/pending/$vm_ID
    exit -1
fi

mkdir -p $image_cache/ocp/$ocp_version/$virt_type
kernel=ocp/$ocp_version/$virt_type/rhcos-installer-kernel
if [ ! -f "$image_cache/$kernel" ]; then
    wget -q $image_repo/$kernel -O $image_cache/$kernel
    if [ ! -f "$image_cache/$kernel" ]; then
        echo "$vm_ID: no kernel file!"
        echo "|:-COMMAND-:| launch_vm.sh '$ID' '$vm_stat' '$SCI_CLIENT_ID' '$vm_ID: no kernel file!'"
        rm -f /tmp/cloudland/pending/$vm_ID
        exit -1
    fi
fi
ramdisk=ocp/$ocp_version/$virt_type/rhcos-installer-initramfs.img
if [ ! -f "$image_cache/$ramdisk" ]; then
    wget -q $image_repo/$ramdisk -O $image_cache/$ramdisk
    if [ ! -f "$image_cache/$ramdisk" ]; then
        echo "$vm_ID: no ramdisk file!"
        echo "|:-COMMAND-:| launch_vm.sh '$ID' '$vm_stat' '$SCI_CLIENT_ID' '$vm_ID: no ramdisk file!'"
        rm -f /tmp/cloudland/pending/$vm_ID
        exit -1
    fi
fi

# create parmfile
mkdir -p $xml_dir/$vm_ID
parmfile=$xml_dir/$vm_ID/parmfile
cat > $parmfile <<EOF
rd.neednet=1 coreos.inst.install_dev=dasda coreos.live.rootfs_url=http://$service:8080/rhcos-rootfs.img coreos.inst.ignition_url=http://$service:8080/ignition/${role}.ign rd.dasd=0.0.0100 ip=$ip_address::$gateway:$netmask:::none nameserver=$service rd.znet=qeth,0.0.1000,0.0.1001,0.0.1002,layer2=1,portno=0 zfcp.allow_lun_scan=0 cio_ignore=all,!condev
EOF

vmur punch -r -u $vm_ID -N RHCOS.KERNEL $image_cache/$kernel
vmur punch -r -u $vm_ID -N TEST.PARM $parmfile
vmur punch -r -u $vm_ID -N RHCOS.INITRD $image_cache/$ramdisk

# start VM
rc=$(curl -s $zvm_service/guests/$vm_ID/action -X POST -d '{"action":"start"}' | jq .rc)
if [ $rc -ne 0 ]; then
    # remove disk and user ?
    echo "$vm_ID: Start VM failed!"
    echo "|:-COMMAND-:| launch_vm.sh '$ID' '$vm_stat' '$SCI_CLIENT_ID' '$vm_ID: Start VM failed!'"
    rm -f /tmp/cloudland/pending/$vm_ID
    exit -1
fi

# reset IPL 100
smcli Image_Definition_Update_DM -T $vm_ID -k "IPL=VDEV=100"

rm -f /tmp/cloudland/pending/$vm_ID

vm_stat=running
echo "|:-COMMAND-:| launch_vm.sh '$ID' '$vm_stat' '$SCI_CLIENT_ID' 'unknown'"
# oc_vm2.sh ????
