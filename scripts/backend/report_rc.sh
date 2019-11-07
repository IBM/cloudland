#!/bin/bash

cd `dirname $0`
source ../cloudrc

exec <&-

cpu=0
total_cpu=$(cat /proc/cpuinfo | grep -c processor)
memory=0
total_memory=$(free | grep 'Mem:' | awk '{print $2}')
disk=0
total_disk=$(df -B 1 $image_dir | tail -1 | awk '{print $4}')
network=0
total_network=0
load=$(w | head -1 | cut -d',' -f5 | cut -d'.' -f1 | xargs)
total_load=0

function probe_arp()
{
    cd /opt/cloudland/cache/router
    for router in *; do
        ID=${router##router-}
        ext_ips=$(sudo ip netns exec $router ip addr show te-$ID | grep 'inet ' | awk '{print $2}')
        for ip in $ext_ips; do
            sudo ip netns exec $router arping -c 1 -I te-$ID ${ip%%/*}
        done
        int_ips=$(sudo ip netns exec $router ip addr show ti-$ID | grep 'inet ' | awk '{print $2}')
        for ip in $int_ips; do
            sudo ip netns exec $router arping -c 1 -I ti-$ID ${ip%%/*}
        done
    done
    cd -
}

function inst_status()
{
    old_inst_list=$(cat $image_dir/old_inst_list)
    inst_list=$(sudo virsh list --all | tail -n +3 | cut -d' ' -f3- | xargs | sed 's/inst-//g;s/shut off/shut_off/g')
    [ "$inst_list" = "$old_inst_list" ] && return
    [ -n "$inst_list" ] && echo "|:-COMMAND-:| inst_status.sh '$SCI_CLIENT_ID' '$inst_list'"
    echo "$inst_list" >$image_dir/old_inst_list
}

function vlan_status()
{
    cd /opt/cloudland/cache/dnsmasq
    old_vlan_list=$(cat old_vlan_list)
    vlan_list=$(ls vlan*)
    vlan_list=$(echo "$vlan_list $(ip netns list | grep vlan | cut -d' ' -f1)" | xargs | sed 's/vlan//g')
    [ "$vlan_list" = "$old_vlan_list" ] && return
    [ -n "$vlan_list" ] && echo "|:-COMMAND-:| vlan_status.sh '$SCI_CLIENT_ID' '$vlan_list'"
    echo "$vlan_list" >old_vlan_list
}

function router_status()
{
    cd /opt/cloudland/cache/router
    old_router_list=$(cat old_router_list)
    router_list=$(ls router*)
    router_list=$(echo "$router_list $(ip netns list | grep router | cut -d' ' -f1)" | xargs | sed 's/router-//g')
    [ "$router_list" = "$old_router_list" ] && return
    [ -n "$router_list" ] && echo "|:-COMMAND-:| router_status.sh '$SCI_CLIENT_ID' '$router_list'"
    echo "$router_list" >old_router_list
}

function calc_resource()
{
    virtual_cpu=0
    virtual_memory=0
    virtual_disk=0
    for xml in $(ls $xml_dir/*/*.xml 2>/dev/null); do
        vcpu=$(xmllint --xpath 'string(/domain/vcpu)' $xml)
        vmem=$(xmllint --xpath 'string(/domain/memory)' $xml)
        [ -z "$vcpu" -o -z "$vmem" ] && continue
        let virtual_cpu=$virtual_cpu+$vcpu
        let virtual_memory=$virtual_memory+$vmem
    done
    used_disk=$(du $image_dir | awk '{print $1}')
    for disk in $(ls $image_dir/* 2>/dev/null); do
        vdisk=$(qemu-img info $disk | grep 'virtual size:' | cut -d' ' -f4 | tr -d '(')
        [ -z "$vdisk" ] && continue
        let virtual_disk=$virtual_disk+$vdisk
    done
    let disk=($total_disk-$used_disk)*$disk_over_ratio-$virtual_disk
    [ $disk -lt 0 ] && disk=0
    let cpu=$total_cpu*$cpu_over_ratio-$virtual_cpu
    [ $cpu -lt 0 ] && cpu=0
    let memory=$total_memory*$mem_over_ratio-$virtual_memory
    [ $memory -lt 0 ] && memory=0
    let total_cpu=$total_cpu*$cpu_over_ratio
    let total_memory=$total_memory*$mem_over_ratio
    let total_disk=($total_disk-$used_disk)*$disk_over_ratio
    echo "cpu=$cpu/$total_cpu memory=$memory/$total_memory disk=$disk/$total_disk network=$network/$total_network load=$load/$total_load"
    echo "|:-COMMAND-:| hyper_status.sh '$SCI_CLIENT_ID' '$HOSTNAME' '$cpu/$total_cpu' '$memory/$total_memory' '$disk/$total_disk'"
}

calc_resource
probe_arp >/dev/null 2>&1
inst_status
vlan_status
router_status
