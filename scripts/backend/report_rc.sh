#!/bin/bash

cd `dirname $0`
source ../cloudrc

exec <&-

cpu=0
total_cpu=$(cat /proc/cpuinfo | grep -c processor)
memory=0
total_memory=$(free | grep 'Mem:' | awk '{print $2}')
disk=0
disk_info=$(df -B 1 $image_dir | tail -1)
total_disk=$(echo $disk_info | awk '{print $2}')
mount_point=$(echo $disk_info | awk '{print $6}')
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
        ext_mac=$(sudo ip netns exec $router ip -o link show te-$ID | awk '{print $17}')
        for ip in $ext_ips; do
            sudo ip netns exec $router arping -c 1 -I te-$ID ${ip%%/*}
            [ -n "$zlayer2_interface" ] && sudo /usr/sbin/bridge fdb add $ext_mac dev $zlayer2_interface
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
    vlan_list=$(ls | grep vlan | grep -v old_vlan_list | xargs | sed 's/vlan//g')
    [ "$vlan_list" = "$old_vlan_list" ] && return
    vlan_arr=($vlan_list)
    nlist=$(ip netns list | grep vlan | cut -d' ' -f1 | xargs | sed 's/vlan//g')
    vlan_status_list=""
    for var in ${vlan_arr[*]}; do
        status="INACTIVE"
        [[ $nlist =~ $var ]] && status="ACTIVE"
        first=""
        [[ -d "vlan$var" ]] && [[ -f "vlan$var/vlan$var.FIRST" ]] && first="FIRST"
        second=""
        [[ -d "vlan$var" ]] && [[ -f "vlan$var/vlan$var.SECOND" ]] && second="SECOND"
        vlan_status_list="$vlan_status_list $var:$status:$first:$second"
    done
    vlan_status_list=$(echo $vlan_status_list | sed -e 's/^[ ]*//g')
    [ -n "$vlan_status_list" ] && echo "|:-COMMAND-:| vlan_status.sh '$SCI_CLIENT_ID' '$vlan_status_list'"
    echo "$vlan_list" >old_vlan_list
}

function router_status()
{
    cd /opt/cloudland/cache/router
    old_router_list=$(cat old_router_list)
    router_list=$(ls router* 2>/dev/null)
    router_list=$(echo "$router_list $(sudo ip netns list | grep router | cut -d' ' -f1)" | xargs | sed 's/router-//g')
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
    used_disk=$(sudo du -s $image_dir | awk '{print $1}')
    for disk in $(ls $image_dir/* 2>/dev/null); do
        vdisk=$(qemu-img info $disk | grep 'virtual size:' | cut -d' ' -f4 | tr -d '(')
        [ -z "$vdisk" ] && continue
        let virtual_disk=$virtual_disk+$vdisk
    done
    total_used_disk=$(sudo du -s $mount_point | awk '{print $1}')
    total_disk=$(echo "($total_disk-$total_used_disk+$used_disk)*$disk_over_ratio" | bc)
    total_disk=${total_disk%.*}
    disk=$(echo "$total_disk-$virtual_disk" | bc)
    disk=${disk%.*}
    [ $disk -lt 0 ] && disk=0
    total_cpu=$(echo "$total_cpu*$cpu_over_ratio" | bc)
    total_cpu=${total_cpu%.*}
    cpu=$(echo "$total_cpu-$virtual_cpu" | bc)
    cpu=${cpu%.*}
    [ $cpu -lt 0 ] && cpu=0
    total_memory=$(echo "$total_memory*$mem_over_ratio" | bc)
    total_memory=${total_memory%.*}
    memory=$(echo "$total_memory-$virtual_memory" | bc)
    memory=${memory%.*}
    [ $memory -lt 0 ] && memory=0
    state=1
    if [ -f "$run_dir/disabled" ]; then
        echo "cpu=0/$total_cpu memory=0/$total_memory disk=0/$total_disk network=$network/$total_network load=$load/$total_load"
        state=0
    else
        echo "cpu=$cpu/$total_cpu memory=$memory/$total_memory disk=$disk/$total_disk network=$network/$total_network load=$load/$total_load"
    fi
    cd /opt/cloudland/run
    old_resource_list=$(cat old_resource_list)
    resource_list="'$cpu' '$total_cpu' '$memory' '$total_memory' '$disk' '$total_disk' '$state'"
    [ "$resource_list" = "$old_resource_list" ] && return
    echo "|:-COMMAND-:| hyper_status.sh '$SCI_CLIENT_ID' '$HOSTNAME' '$cpu' '$total_cpu' '$memory' '$total_memory' '$disk' '$total_disk' '$state'"
    echo "'$cpu' '$total_cpu' '$memory' '$total_memory' '$disk' '$total_disk' '$state'" >/opt/cloudland/run/old_resource_list
}

function replace_vnc_passwd()
{
    old_timestamp=/opt/cloudland/run/last_vnc_update
    [ ! -f $old_timestamp ] && touch $old_timestamp && return
    time_stamp=$(stat -c %X $old_timestamp)
    let duration=$(date +%s)-$time_stamp
    if [ "$duration" -gt 300 ]; then 
        for inst in $(sudo virsh list --all | grep inst | awk '{print $2}'); do
            sudo ./replace_vnc_passwd.sh $inst
        done
        touch $old_timestamp
    fi
}

replace_vnc_passwd
calc_resource
probe_arp >/dev/null 2>&1
inst_status
vlan_status
router_status
