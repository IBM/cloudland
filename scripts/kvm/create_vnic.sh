#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 11 ] && die "$0 <vm_ID> <image> <vlan> <mac> <name> <ip> <netmask> <gateway> <cpu> <memory> <disk_inc> [userdata] [pubkey]"

vm_ID=$1
img_name=$2
vlan=$3
vm_mac=$4
vm_name=$5
vm_ip=$6 
vm_mask=$7
vm_gw=$8
vm_cpu=$9
vm_mem=${10}
disk_inc=${11}
userdata=${12}
pubkey=${13}
vm_stat=error
vm_vnc=""
use_lb=false

instance=$vm_ID
ip netns add $instance
suffix=$(echo $vm_mac | cut -d: -f3- | tr -d ':')
vs_nic=tap$suffix
vm_nic=vm-$suffix
ip link add dev $vs_nic type veth peer name $vm_nic
ip link set $vs_nic up
ip link set $vm_nic netns $instance
ip netns exec $instance ip link set $vm_nic up
prefix=$(ipcalc -p $vm_ip $vm_mask | cut -d= -f2)
brdcast=$(ipcalc -b $vm_ip $vm_mask | cut -d= -f2)
ip netns exec $instance ip link set address $vm_mac dev $vm_nic
ip netns exec $instance ip link set $vm_nic mtu 1450
ip netns exec $instance ip addr add ${vm_ip}/$prefix brd $brdcast dev $vm_nic
ip netns exec $instance route add default gw $vm_gw
ip netns exec $instance ip link set lo up
hyper_ip=$(ifconfig $vxlan_interface | grep 'inet addr:' | cut -d: -f2 | cut -d' ' -f1)
br_name=br$SCI_CLIENT_ID
tunip=$(inet_aton 172.250.0.10)
let tunip=$tunip+$SCI_CLIENT_ID
tunip=$(inet_ntoa $tunip)
ip addr add $tunip/16 brd 172.250.255.255 dev $vxlan_interface
ovs-vsctl --may-exist add-br $br_name
cmd="icp-tower --ovs-bridge=$br_name gate add --direct-routing --encap-identifier $vlan --local-ip=$tunip --interface $vs_nic --vsi-mac-address $vm_mac --vsi-ip-prefix ${vm_ip}/$prefix --decapper-ip $decapper_ip"
result=$(eval "$cmd")
./add_fdb.sh &
sidecar span log $span "Callback: create_vnic.sh '$vs_nic' '$SCI_CLIENT_ID' '$vlan' '$vm_ip/$prefix' '$vm_mac' '$tunip'"
echo "|:-COMMAND-:| create_vnic.sh '$vs_nic' '$SCI_CLIENT_ID' '$vlan' '$vm_ip/$prefix' '$vm_mac' '$tunip'"
sidecar span log $span "Callback: `basename $0` 'running' '$SCI_CLIENT_ID' '$vm_mac' '$vm_vnc' '$vsize' '$is_vol'"
echo "|:-COMMAND-:| `basename $0` '$vm_ID' 'running' '$SCI_CLIENT_ID' '$vm_mac' '$vm_vnc' '$vsize' '$is_vol'"
if [ -n "$userdata" ]; then
    usersh=$cache_dir/meta/$vm_ID.sh
    echo "$userdata" > $usersh
    chmod +x $usersh
    start-stop-daemon -S -b -x $usersh $vm_ID $vm_ip $vm_mask
fi
