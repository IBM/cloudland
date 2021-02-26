#!/bin/bash

cd `dirname $0`
source ../cloudrc

exec <&-

function calc_resource()
{
    virtual_cpu=0
    virtual_memory=0

    guest_output=$(curl -s $zvm_service/guests | jq .output)
    num_guest=$(echo $guest_output | jq length)
    for((i=0;i<$num_guest;i++));  do
       guest=$(echo $guest_output | jq .[$i] | tr -d '"')
       guest_info=$(curl -s $zvm_service/guests/$guest/info | jq .output)
       vcpu=$(echo $guest_info | jq .num_cpu)
       vmem=$(echo $guest_info | jq .max_mem_kb)
       let virtual_cpu=$virtual_cpu+$vcpu
       let virtual_memory=$virtual_memory+$vmem/1024
    done

    host_info=$(curl -s $zvm_service/host | jq .output)
    disk_available=$(echo $host_info | jq .disk_available)
    disk_total=$(echo $host_info | jq .disk_total)
    memory_total=$(echo $host_info | jq .memory_mb)

    cpu_total=$(curl -s $zvm_service/host | jq .output.vcpus)
    let cpu_total=$cpu_total*$cpu_over_ratio
    cpu=$virtual_cpu
    let cpu=$cpu_total-$cpu
    [ $cpu -lt 0 ] && cpu=0

    let memory_available=$memory_total-$virtual_memory
    [ $memory_available -lt 0 ] && memory_available=0
    let memory_available=memory_available*1024
    let memory_total=memory_total*1024
    let disk_available=disk_available*1024*1024*1024
    let disk_total=disk_total*1024*1024*1024

    state=1
    if [ -f "$run_dir/disabled" ]; then
        echo "cpu=0/$cpu_total memory=0/$memory_total disk=0/$disk_total"
        state=0
    else
        echo "cpu=$cpu/$cpu_total memory=$memory_available/$memory_total disk=$disk_available/$disk_total"
    fi
    cd /opt/cloudland/run
    old_resource_list=$(cat old_resource_list 2>/dev/null)
    resource_list="'$cpu' '$cpu_total' '$memory_available' '$memory_total' '$disk_available' '$disk_total' '$state'"
    [ "$resource_list" = "$old_resource_list" ] && return
    echo "|:-COMMAND-:| hyper_status.sh '$SCI_CLIENT_ID' '$HOSTNAME' '$cpu' '$cpu_total' '$memory_available' '$memory_total' '$disk_available' '$disk_total' '$state' '$VIRT_TYPE' '$ZONE_NAME'"
    echo "'$cpu' '$cpu_total' '$memory_available' '$memory_total' '$disk_available' '$disk_total' '$state'" >/opt/cloudland/run/old_resource_list
}

function inst_status()
{
    old_inst_list=$(cat $image_dir/old_inst_list 2>/dev/null)
    inst_list=""
    guest_output=$(curl -s $zvm_service/guests | jq .output)
    num_guest=$(echo $guest_output | jq length)
    for((i=0;i<$num_guest;i++));  do
        guest=$(echo $guest_output | jq .[$i] | tr -d '"')
        if [ -e /tmp/cloudland/pending/$guest ]; then
            continue
        fi
        ID=$(echo ${guest:3:5})
        ID="0x$ID"
        ID=$(printf "%d" $ID)
        state=$(curl -s $zvm_service/guests/$guest/power_state | jq .output | tr -d '"')
        if [ $state = "on"  ]; then
            inst_list="${inst_list} $ID running"
        else
            inst_list="${inst_list} $ID shut_off"
        fi
    done
    inst_list=$(echo $inst_list | sed 's/^[ \t]*//g')
    [ "$inst_list" = "$old_inst_list" ] && return
    [ -n "$inst_list" ] && echo "|:-COMMAND-:| inst_status.sh '$SCI_CLIENT_ID' '$inst_list'"
    echo "$inst_list" >$image_dir/old_inst_list
}

calc_resource
inst_status
