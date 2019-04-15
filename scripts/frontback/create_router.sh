#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <router> <hyper>" && exit -1

router=$1
hyper=$2

sql_exec "update router set host='$hyper' where name='$router'"

