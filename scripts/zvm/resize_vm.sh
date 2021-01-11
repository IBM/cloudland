#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 6 ] && die "$0 <vm_ID> <cpu> <memory> <disk_size> <swap_size> <ephemeral_size> <increased_disk_size>"

ID=$1
vm_ID=$(printf $guest_userid_template $1)
vm_cpu=$2
vm_mem=$3
disk_size=$4
swap_size=$5
ephemeral_size=$6
increased_disk_size=$7

# Stop VM
rc=$(curl -s $zvm_service/guests/$vm_ID/action -X POST -d '{"action":"softstop"}' | jq .rc)
if [ $rc -ne 0 ]; then
   echo "$vm_ID stop failed."
fi

# CPU
rc=$(smcli Image_Definition_Update_DM -T $vm_ID -k "CPU_MAXIMUM=COUNT=$vm_cpu TYPE=ESA")
if [ $rc != "Done" ]; then
    echo "$vm_ID set CPU failed."
fi

rc=$(curl -s $zvm_service/guests/$vm_ID/action -X POST -d '{"action":"resize_cpus", "cpu_cnt": '$vm_cpu'}' | jq .rc)
if [ $rc -ne 0 ]; then
    echo "$vm_ID resize CPU failed."
fi

# Storage(Memory)
rc=$(smcli Image_Definition_Update_DM -T $vm_ID -k "STORAGE_MAXIMUM=${vm_mem}M")
if [ $rc != "Done" ]; then
    echo "$vm_ID set memory max failed."
fi

rc=$(smcli Image_Definition_Update_DM -T $vm_ID -k "STORAGE_INITIAL=${vm_mem}M")
if [ $rc != "Done" ]; then
    echo "$vm_ID set memory initial failed."
fi

# Disk
if [ $increased_disk_size -gt 0 ]; then
    rc=$(curl -s $zvm_service/guests/$vm_ID/disks -X POST -d '{"disk_info":{"disk_list":[{"size":"'"$increased_disk_size"'G", "is_boot_disk":"False"}]}}' | jq .rc)
    if [ $rc -ne 0 ]; then
        echo "$vm_ID change disk failed."
    fi
fi

# Start VM
rc=$(curl -s $zvm_service/guests/$vm_ID/action -X POST -d '{"action":"start"}' | jq .rc)
if [ $rc -ne 0 ]; then
   echo "$vm_ID start VM failed."
fi

state=$(curl -s $zvm_service/guests/$vm_ID/power_state | jq .output | tr -d '"')
if [ $state = "on"  ]; then
    echo "|:-COMMAND-:| inst_status.sh '$SCI_CLIENT_ID' '$ID running'"
else
    echo "|:-COMMAND-:| inst_status.sh '$SCI_CLIENT_ID' '$ID shut_off'"
fi

