cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && echo "$0 <vm_ID> <nic_name> <inbound> <outbound>" && exit -1

ID=$1
nic_name=$2
inbound=$3
outbound=$4

[ -z "$inbound" -o "$inbound" -eq 0 ] && inbound=1000
inbound_burst=$inbound
inbound_rate=$(( $inbound * 1000 ))
[ -z "$outbound" -o "$outbound" -eq 0 ] && outbound=1000
outbound_burst=$outbound
outbound_rate=$(( $outbound * 1000 ))
virsh domiftune $vm_ID $nic_name --inbound $inbound_rate,$inbound_rate,$inbound_burst --outbound $outbound_rate,$outbound_rate,$outbound_burst
