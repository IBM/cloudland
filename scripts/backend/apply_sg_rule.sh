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
        apply_fw $action $chain -p $proto $args -j RETURN
    elif [ "$max" -eq "$min" ]; then
        apply_fw $action $chain -p $proto -m $proto --dport $max $args -j RETURN
    elif ["$max" -gt "$min" ]; then
        apply_fw $action $chain -p $proto -m $proto -m multiport --dport $min:$max $args -j RETURN
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

while read line; do
    [ -z "$line" ] && continue
    direction=$(echo $line | cut -d' ' -f1 | xargs)
    remote=$(echo $line | cut -d' ' -f2 | xargs)
    protocol=$(echo $line | cut -d' ' -f3 | xargs)
    chain=$chain_in
    [ "$direction" = "egress" ] && chain=$chain_out
    if [ -n "$remote" ]; then
        [ "$direction" = "ingress" ] && args="-s $remote"
        [ "$direction" = "egress" ] && args="-d $remote"
    fi
    port_min=$(echo $line | cut -d' ' -f4 | xargs)
    port_max=$(echo $line | cut -d' ' -f5 | xargs)
    case "$protocol" in
        "TCP")
            allow_ipv4 "$chain" "$args" "tcp" "$port_min" "$port_max"
            ;;
        "UDP")
            allow_ipv4 "$chain" "$args" "udp" "$port_min" "$port_max"
            ;;
        "ICMP")
            ptype=$port_min
            pcode=$port_max
            allow_icmp "$chain" "$args" "$ptype" "$pcode"
            ;;
        *)
            apply_fw "$action" "$chain" "-p" "$protocol" "$args" -j RETURN
            ;;
    esac
done
