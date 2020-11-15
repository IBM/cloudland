#!/bin/bash

vm_ID=$1
ip=$2
mask=$3

exec >>/tmp/ping.log 2>&1

function inet_aton()
{   
    ip="$1"
    hex=`printf '%02x' ${ip//./ }`
    printf "%lu\n" "0x${hex}"
}

function inet_ntoa()
{   
    num="$1"
    hex=`printf "%08x\n" ${num}`
    for i in `echo ${hex} | sed "s/\(..\)/\1 /g"`; do
        printf '%hu.' "0x${i}"
    done | sed "s/\.$//g"
}

hostmin=$(ipcalc $ip $mask | grep HostMin | awk '{print $2}')
let pingip=$(inet_aton $hostmin)+1
#myip=$(inet_aton $ip)

addr=$(inet_ntoa $pingip)
ip netns exec $vm_ID ping -c 1 -w 3 $addr
[ $? -eq 0 ] && curl http://10.171.202.186:30080/metrics/cland/server/number?state=running

#success=0
#while [ "$pingip" -lt "$myip" ]; do
#    let pingip=$pingip+1
#    addr=$(inet_ntoa $pingip)
#    ip netns exec $vm_ID ping -c 1 -w 3 $addr
#    [ $? -eq 0 ] && let success=$success+1
#done

#echo $success
