#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <user>" && exit -1

owner=$1
sql_exec "select vlan, description, owner, router from netlink where owner='$owner' or shared='true' COLLATE NOCASE"
