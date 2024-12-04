#!/bin/bash

cd `dirname $0`
source ../cloudrc

usera=$1
userb=$2

sql_exec "update users set role='$role' where username='$username'"
export usera=jiajj@cn.ibm.com
export userb=diaojuan@cn.ibm.com
sql_exec "update instance set owner='$usera' where owner='$userb'"
