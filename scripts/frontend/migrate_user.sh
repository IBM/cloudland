#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <original_user> <bluepage_id>" && exit -1

org_user=$1
blue_id=$2

sql_exec "update image set owner='$2' where owner='$1'"
sql_exec "update instance set owner='$2' where owner='$1'"     
sql_exec "update netlink set owner='$2' where owner='$1'"
sql_exec "update snapshot set owner='$2' where owner='$1'"
sql_exec "update volume set owner='$2' where owner='$1'"      
