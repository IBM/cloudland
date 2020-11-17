#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <interface> [add|delete]" && exit -1

vnic=$1
act=$2
action='-I'
[ "$act" = "delete" ] && action='-D'

chain_in=secgroup-in-$vnic
chain_out=secgroup-out-$vnic

function allow_ipv4()
{
    chain=$1
    args=$2
    proto=$3
    min=$4
    max=$5
    if [ -z "$min" -a -z "$max" ]; then
        apply_fw $action $chain -p $proto $args -m conntrack --ctstate NEW -j RETURN
    elif [ "$max" -eq "$min" ]; then
        apply_fw $action $chain -p $proto -m $proto -m conntrack --ctstate NEW --dport $max $args -j RETURN
    elif [ "$max" -gt "$min" ]; then
        apply_fw $action $chain -p $proto -m $proto -m conntrack --ctstate NEW --dport $min:$max $args -j RETURN
    fi
}

function allow_icmp()
{
    chain=$1
    args=$2
    ptype=$3
    pcode=$4
    if [ "$ptype" != "-1" ]; then
        typecode=$ptype
        [ "$pcode" != "-1" ] && typecode=$ptype/$pcode
        args="$args --icmp-type $typecode"
    fi
    apply_fw $action $chain -p icmp $args -j RETURN
}

sec_data=$(cat)
i=0
len=$(jq length <<< $sec_data)
while [ $i -lt $len ]; do
    direction=$(jq -r .[$i].direction <<< $sec_data)
    remote_ip=$(jq -r .[$i].remote_ip <<< $sec_data)
    protocol=$(jq -r .[$i].protocol <<< $sec_data)
    chain=$chain_in
    [ "$direction" = "egress" ] && chain=$chain_out
    if [ -n "$remote_ip" ]; then
        [ "$direction" = "ingress" ] && args="-s $remote_ip"
        [ "$direction" = "egress" ] && args="-d $remote_ip"
    fi
    port_min=$(jq -r .[$i].port_min <<< $sec_data)
    port_max=$(jq -r .[$i].port_max <<< $sec_data)
    case "$protocol" in
        "tcp")
            allow_ipv4 "$chain" "$args" "tcp" "$port_min" "$port_max"
            ;;
        "udp")
            allow_ipv4 "$chain" "$args" "udp" "$port_min" "$port_max"
            ;;
        "icmp")
            ptype=$port_min
            pcode=$port_max
            allow_icmp "$chain" "$args" "$ptype" "$pcode"
            ;;
        *)
            apply_fw "$action" "$chain" "-p" "$protocol" "$args" -j RETURN
            ;;
    esac
    let i=$i+1
done

service iptables save
