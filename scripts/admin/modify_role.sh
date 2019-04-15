#!/bin/bash

cd `dirname $0`
source ../cloudrc

case "$2" in
    "admin" | "user" | "vip" ) true;;
    * ) false;;
esac

[ $? != 0 -o $# -lt 2 ] && echo "$0 <username> <admin|user|vip>" && exit -1

username=$1
role=$2

sql_exec "update users set role='$role' where username='$username'"
