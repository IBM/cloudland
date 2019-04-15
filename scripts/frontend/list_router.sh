#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <user>" && exit -1

owner=$1
sql_exec "select name, description, out_addr, vlans, owner from router where owner='$owner' or shared='true' COLLATE NOCASE"
