#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <vlan> <hyper>" && exit -1

vlan=$1
hyper=$2

sql_exec "update netlink set dh_host='$hyper' where vlan='$vlan'"

