#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <router> <ext_vlan>" && exit -1

ID=$1
router=router-$ID
ext_vlan=$2

rtID=$(( $ext_vlan % 250 + 2 ))
table=fip-$ext_vlan
rt_file=/etc/iproute2/rt_tables
grep "^$rtID $table" $rt_file
if [ $? -ne 0 ]; then
    for i in {1..250}; do
        grep "^$rtID\s" $rt_file
	[ $? -ne 0 ] && break
        rtID=$(( ($rtID + 17) % 250 + 2 ))
    done
    echo "$rtID $table" >>$rt_file
fi
suffix=${ID}-${ext_vlan}
./create_veth.sh $router ext-$suffix te-$suffix
