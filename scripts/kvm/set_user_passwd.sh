#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && die "$0 <vm_ID> <user> <passwd>"

vm_ID=inst-$1
username=$2
passwd=$3
timeout_virsh set-user-password --domain $vm_ID --user $username --password $passwd
[ $? -ne 0 ] && die "Failed to set user password"
echo "|:-COMMAND-:| $(basename $0) '$1' 'success'"
