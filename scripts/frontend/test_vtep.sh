#!/bin/bash -xv

cd `dirname $0`

FLIGHT_PLANNER_ENDPOINT="http://192.168.1.1:5000"
VPC_NAME="compute_demo"

NETWORK="172.16.11.0"
NETMASK="255.255.255.0"

id=$(sqlite3 /opt/cloudland/db/cloudland.db "select max(id)+1 from vtep")
instance=vtep$id
vni=$(curl ${FLIGHT_PLANNER_ENDPOINT}/encaps/${VPC_NAME} | jq .vni)
./list_link.sh admin | grep -q "\<$vni\>"
if [ $? -ne 0 ]; then
    ./create_vlan.sh admin $vni $VPC_NAME
    ./create_net.sh admin $vni $NETWORK $NETMASK
fi
./create_vtep.sh $instance $vni
