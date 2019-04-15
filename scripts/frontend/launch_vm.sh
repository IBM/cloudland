#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <user> <image> <vlan> [mac] [name] [ip] [cpu] [memory(m)] [disk_inc(G)] [hyper] [userdata] [pubkey]" && exit -1

owner=$1
img_name=$2
vlan=$3
vm_mac=$4
vm_name=$5
vm_ip=$6
vm_cpu=$7
vm_mem=$8
disk_inc=$9
hyper=${10}
userdata=${11}
pubkey=${12}
vm_ID=`date +%m%d%H%M%S-%N`
vm_mem=${vm_mem%[m|M]}

if [ -f "$volume_dir/$img_name.disk" ]; then
    vm_ID=$img_name
    num=`sql_exec "select count(*) from instance where inst_id='$vm_ID' and status!='deleted'"`
    [ $num -gt 0 ] && die "Existing instance is using volume $vm_ID!"
fi
if [ -n "$disk_inc" ]; then
    disk_inc=${disk_inc%%[G|g]} 
    [ $disk_inc -gt 0 -a $disk_inc -le $disk_inc_limit ] || die "Invalid disk increase size $disk_inc!"
fi
[ -z "$vm_cpu" ] && vm_cpu=1
[ -z "$vm_mem" ] && vm_mem=256
[ $vm_cpu -le $cpu_limit -a $vm_cpu -gt 0 ] || die "Valid cpu number is 1 - $cpu_limit!"
[ $vm_mem -le $mem_limit -a $vm_mem -gt 128 ] || die "Valid memory is 128 - $mem_limit!"
num=`sql_exec "select count(*) from instance where owner='$owner' and status!='deleted'"`
quota=`sql_exec "select inst_limit from quota where role=(select role from users where username='$owner')"`
[ $quota -ge 0 -a $num -ge $quota ] && die "Quota is used up!"
num=`sql_exec "select count(*) from netlink where vlan='$vlan' and (owner='$owner' or shared='true' COLLATE NOCASE)"`
[ $num -lt 1 ] && die "Not authorised to launch vm on vlan $vlan!"

if [ -n "$vm_ip" ]; then
    allocated=`sql_exec "select IP from address where vlan='$vlan' and IP='$vm_ip'"`
    [ "$allocated" != "false" ] && die "IP $vm_ip is not available!"
fi
[ -z "$vm_ip" ] && vm_ip=`sql_exec "select IP from address where vlan='$vlan' and allocated='false' limit 1"`
[ -z "$vm_ip" ] && die "No IP address is avalable"
sql_exec "update address set allocated='true' where IP='$vm_ip'"
[ -z "$vm_name" ] && vm_name=HOST-`echo $vm_ip | tr '.' '-'`
vm_name=`echo $vm_name | sed -e 's/[^-A-Za-z0-9]//g' -e 's/^[-0-9]*//g' -e 's/[-]*$//g'`

dns_host=/opt/cloudland/dnsmasq/vlan$vlan.host
dns_opt=/opt/cloudland/dnsmasq/vlan$vlan.opts
[ -z "$vm_mac" ] && vm_mac="52:54:"`openssl rand -hex 4 | sed 's/\(..\)/\1:/g; s/.$//'`

sql_exec "insert into instance (inst_id, hname, vlan, mac_addr, ip_addr, owner, status, image, cpu, memory) values ('$vm_ID', '$vm_name', '$vlan', '$vm_mac', '$vm_ip', '$owner', 'launching', '$img_name', '$vm_cpu', '$vm_mem')"
sql_exec "update address set allocated='true', mac='$vm_mac', instance='$vm_ID' where IP='$vm_ip'"
hyper_id=`sql_exec "select id from compute where hyper_name='$hyper'"`
dh_host=`sql_exec "select id from compute where hyper_name=(select dh_host from netlink where vlan='$vlan')"`
[ "$dh_host" -ge 0 ] && /opt/cloudland/bin/sendmsg "inter $dh_host" "/opt/cloudland/scripts/backend/set_host.sh $vlan $vm_mac $vm_name $vm_ip"
#./build_meta.sh "$userdata" "$pubkey" "$vm_ID" "$vm_name"
/opt/cloudland/bin/sendmsg "inter $hyper_id" "/opt/cloudland/scripts/backend/`basename $0` '$vm_ID' '$img_name' '$vlan' '$vm_mac' '$vm_name' '$vm_ip' '$vm_cpu' '$vm_mem' '$disk_inc'"
echo "$vm_ID|launching"
