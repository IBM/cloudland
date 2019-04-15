#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <router> <vlan> <gateway>" && exit -1

router=$1
vlan=$2
gateway=$3

sql_exec "update network set router='$router' where vlan='$vlan' and gateway='$gateway'"
