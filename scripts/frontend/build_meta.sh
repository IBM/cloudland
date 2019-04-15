#!/bin/bash
cd `dirname $0`
source ../cloudrc
[ $# -lt 4 ] && echo "$0 <userdata> <pubkey> <vm_ID> <vm_name>" && exit -1
userdata=$1
pubkey=$2
vm_ID=$3
vm_name=$4
working_dir=/tmp/$vm_ID
latest_dir=$working_dir/openstack/latest
mkdir -p $latest_dir
if [ -f "$userdata" ]; then
   cp -f $userdata $latest_dir/user_data
fi
if [ -f "$pubkey" ]; then
    pubkey=`cat $pubkey`
fi
admin_pass=`openssl rand -base64 12`
random_seed=`cat /dev/urandom | head -c 512 | base64 -w 0`
(
    echo '{'
    echo '  "name": "'${vm_name}'",'
    if [ -n "${pubkey}" ]; then
        echo '  "public_keys": {"cloudland": "'${pubkey}'\n"},'
    fi
    echo '  "launch_index": 0,'
    echo '  "hostname": "'${vm_name}'",'
    echo '  "availability_zone": "cloudland",'
    echo '  "uuid": "'${vm_ID}'",'
    echo '  "admin_pass": "'${admin_pass}'",'
    echo '  "random_seed": "'${random_seed}'"'
    echo '}'
) > $latest_dir/meta_data.json
mkisofs -quiet -R -V config-2 -o ${cache_dir}/meta/${vm_ID}.iso $working_dir &> /dev/null
rm -rf $working_dir
