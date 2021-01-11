#!/bin/bash
cd `dirname $0`
source ../cloudrc
[ $# -lt 2 ] && echo "$0 <vm_ID> <vm_name>" && exit -1

vm_ID=$1
vm_name=$2
[ "${vm_name%%.*}" = "$vm_name" ] && vm_name=${vm_name}.$cloud_domain
working_dir=/tmp/cloudland/$vm_ID
latest_dir=$working_dir/openstack/latest
mkdir -p $latest_dir

vm_meta=$(cat)
userdata=$(echo $vm_meta | base64 -d | jq -r .userdata)
if [ -n "$userdata" ]; then
   #echo "$userdata" > $latest_dir/user_data
   echo $vm_meta | base64 -d | jq -r .userdata > $latest_dir/user_data
fi

pub_keys=$(echo $vm_meta | base64 -d | jq -r '.keys')
admin_pass=`openssl rand -base64 12`
random_seed=`cat /dev/urandom | head -c 512 | base64 -w 0`
(
    echo '{'
    echo '  "name": "'${vm_name}'",'
    if [ -n "${pub_keys}" ]; then
        echo -n '  "public_keys": {'
        i=0
        n=$(jq length <<< $pub_keys)
        while [ $i -lt $n ]; do
            key=$(jq -r .[$i] <<< $pub_keys)
            [ $i -ne 0 ] && echo -n ','
            echo -n '"key'$i'": "'$key'\n"'
            let i=$i+1
        done
        echo '},'
    fi
    echo '  "launch_index": 0,'
    echo '  "hostname": "'${vm_name}'",'
    echo '  "availability_zone": "cloudland",'
    echo '  "uuid": "'${vm_ID}'",'
    echo '  "admin_pass": "'${admin_pass}'",'
    echo '  "random_seed": "'${random_seed}'"'
    echo '}'
) > $latest_dir/meta_data.json

# vm_meta=$(echo $vm_meta | base64 -d)
# net_json=$(jq 'del(.userdata) | del(.vswitch) | del(.osVersion) | del(.diskType) | del(.virt_type) | del(.dns) | del(.vlans) | del(.keys) | del(.security)' <<< $vm_meta | jq --arg dns $dns_server '.services[0].type = "dns" | .services[0].address |= .+$dns')


mkdir -p ${cache_dir}/meta/${vm_ID}
mkisofs -quiet -R -V config-2 -o ${cache_dir}/meta/${vm_ID}/cfgdrive.iso  $working_dir &> /dev/null
