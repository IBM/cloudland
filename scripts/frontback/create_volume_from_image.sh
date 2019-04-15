#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <name> <size> <status>" && exit -1

vol_name=$1
vol_size=${2%%[g|G]}
vol_stat=$3

sql_exec "update volume set size='$vol_size', status='$vol_stat' where name='$vol_name'"
