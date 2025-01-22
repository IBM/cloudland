#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && echo "$0 <vm_ID> <nic_name> <inbound> <outbound>" && exit -1

vm_ID=inst-$1
nic_name=$2
inbound=$3
outbound=$4

[ -z "$inbound" -o "$inbound" -eq 0 ] && inbound=1000
inbound_burst=$(( $inbound / 8 ))
inbound_rate=$(( $inbound * 125 )) # in kilobytes per second
inbound_peak=$(( $inbound_rate * 2 ))
#inbound_burst=$(( $inbound_rate * 3 ))
[ -z "$outbound" -o "$outbound" -eq 0 ] && outbound=1000
outbound_rate=$(( $outbound * 125 )) # in kilobytes per second
outbound_peak=$(( $outbound_rate * 2 ))
#outbound_burst=$outbound_peak
virsh domiftune $vm_ID $nic_name --inbound $inbound_rate,$inbound_peak --outbound $outbound_rate,$outbound_peak --config --live
