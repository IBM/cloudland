#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 5 ] && die "$0 <vm_ID> <cpu> <memory> <disk_size> <hostname>"

ID=$1
vm_ID=inst-$1
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
vm_disk=$image_dir/$vm_ID.disk
rm -f $vm_disk
qemu-img create $vm_disk -f qcow2 "${disk_size}G"
metadata=$(cat)
ocp_version=$(jq .ocp[0].ocpVersion <<< $metadata | tr -d '"')
virt_type=$(jq .virt_type <<< $metadata | tr -d '"')
mkdir -p $image_cache/ocp/$ocp_version/$virt_type
kernel=ocp/$ocp_version/$virt_type/rhcos-installer-kernel
ramdisk=ocp/$ocp_version/$virt_type/rhcos-installer-initramfs.img
if [ ! -f "$image_cache/$kernel" ]; then
    wget -q $image_repo/$kernel -O $image_cache/$kernel
fi
if [ ! -f "$image_cache/$ramdisk" ]; then
    wget -q $image_repo/$ramdisk -O $image_cache/$ramdisk
fi
#metadata=$(cat)
echo $metadata > /tmp/cloudland_meta.log
[ -z "$vm_mem" ] && vm_mem='1024m'
[ -z "$vm_cpu" ] && vm_cpu=1
let vm_mem=${vm_mem%[m|M]}*1024
mkdir -p $xml_dir/$vm_ID
vm_xml=$xml_dir/$vm_ID/${vm_ID}.xml
template=$template_dir/openshift.xml
[ $(uname -m) = s390x ] && template=$template_dir/ocd_linux1.xml
cp $template $vm_xml
vlans=$(jq .vlans <<< $metadata)
echo $vlans > /tmp/cloudland_vlan.log
core_ip=$(jq -r .[0].ip_address <<< $vlans)
gw_ip=$(jq .networks[0].routes[0].gateway <<< $metadata | tr -d '"')
lb_ip=$(jq .ocp[0].service <<< $metadata | tr -d '"')
vxlink=v-$(jq -r .[0].vlan <<< $vlans)
mtu_size=1500
[ -d "/sys/devices/virtual/net/$vxlink" ] && mtu_size=1450
ocp_kernel=$image_cache/$kernel
ocp_initramfs=$image_cache/$ramdisk
mkdir -p /tmp/cloudland/xml
sed -i "s/VM_ID/$vm_ID/g; s/VM_MEM/$vm_mem/g; s/VM_CPU/$vm_cpu/g; s#OCP_KERNEL#$ocp_kernel#g; s#OCP_INITRAMFS#$ocp_initramfs#g; s#VM_IMG#$vm_disk#g; s/CORE_IP/$core_ip/g;s/GATEWAY/$gw_ip/g; s/MTU_SIZE/$mtu_size/g; s/HOSTNAME/$hname/g; s/LB_IP/$lb_ip/g; s/ROLE_IGN/${role}.ign/g;" $vm_xml
cp $vm_xml /tmp/cloudland/xml/back_${vm_ID}.xml
state=error
virsh define $vm_xml
nvlan=$(jq length <<< $vlans)
i=0
while [ $i -lt $nvlan ]; do
    vlan=$(jq -r .[$i].vlan <<< $vlans)
    ip=$(jq -r .[$i].ip_address <<< $vlans)
    mac=$(jq -r .[$i].mac_address <<< $vlans)
    jq .security <<< $metadata | ./attach_nic.sh $ID $vlan $ip $mac
    let i=$i+1
done
#brctl addbr brfake
nmcli con add type bridge ifname brfake
virsh attach-interface $vm_ID bridge brfake --model virtio --mac 52:54:11:22:33:44 --config
virsh start $vm_ID
[ $? -eq 0 ] && state=running && ./replace_vnc_passwd.sh $ID
echo "|:-COMMAND-:| launch_vm.sh '$ID' '$state' '$SCI_CLIENT_ID' 'unknown'"
echo "$(pwd)/oc_vm_2.sh '$ID'" | at -t $(date --date="now +20 seconds" +"%Y%m%d%H%M.%S")
