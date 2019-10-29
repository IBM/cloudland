#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <email> <password>" && exit -1

email=$1
echo $email | grep ".*@.*ibm.com" >/dev/null 2>&1
[ $? -ne 0 ] && echo "Invalid email format" && exit 1
passwd="$2"
bind_dn=`ldapsearch -x -b ou=bluepages,o=ibm.com -h bluepages.ibm.com mail=$email | grep dn: | head -1 | cut -d':' -f2`
ldapsearch -x -b ou=bluepages,o=ibm.com -h bluepages.ibm.com mail=$email -D $bind_dn -w "$passwd" >/dev/null 2>&1
