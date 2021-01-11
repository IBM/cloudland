#!/bin/bash

cd `dirname $0`
source ../cloudrc
[ $# -lt 1 ] && die "$0 <vm_ID>"
vm_ID=$1
echo "del_fdb" >> /tmp/del_fdb.log
macid=`cat $xml_dir/${vm_ID}_mac.log`
echo $macid >> /tmp/del_fdb.log
sudo /usr/sbin/bridge fdb del $macid dev $zlayer2_interface
sudo /usr/sbin/bridge fdb show | grep $zlayer2_interface | grep $macid >> /tmp/del_fdb.log
rm -rf $xml_dir/${vm_ID}_mac.log
echo "fdb vlan $vlan records removed" >> /tmp/del_fdb.log