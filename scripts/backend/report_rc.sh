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

function calc_resource()
{
    virtual_cpu=0
    virtual_memory=0
    virtual_disk=0
    for xml in $(ls $xml_dir/*/*.xml 2>/dev/null); do
       vcpu=$(xmllint --xpath 'string(/domain/vcpu)' $xml)
       vmem=$(xmllint --xpath 'string(/domain/memory)' $xml)
       let virtual_cpu=$virtual_cpu+$vcpu
       let virtual_memory=$virtual_memory+$vmem
    done
    used_disk=$(du $image_dir | awk '{print $1}')
    for disk in $(ls $image_dir/* 2>/dev/null); do
        vdisk=$(qemu-img info $disk | grep 'virtual size:' | cut -d' ' -f4 | tr -d '(')
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
}

calc_resource
echo "cpu=$cpu/$total_cpu memory=$memory/$total_memory disk=$disk/$total_disk network=$network/$total_network load=$load/$total_load"
