#!/bin/bash
cd `dirname $0`
source ../cloudrc
[ $# -lt 2 ] && echo "$0 <vm_ID> <vm_name>" && exit -1

vm_ID=$1
vm_name=$2
[ "${vm_name%%.*}" = "$vm_name" ] && vm_name=${vm_name}.$cloud_domain
working_dir=/tmp/$vm_ID
latest_dir=$working_dir/openstack/latest
mkdir -p $latest_dir

vm_meta=$(cat)
userdata=$(echo $vm_meta | base64 -d | jq -r .userdata)
if [ -n "$userdata" ]; then
   echo $vm_meta | base64 -d | jq -r .userdata > $latest_dir/user_data
fi
vm_meta=$(echo $vm_meta | base64 -d)

pub_keys=$(jq -r '.keys' <<< $vm_meta)
root_passwd=$(jq -r '.root_passwd' <<< $vm_meta)

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

if [ -n "${root_passwd}" ]; then
    (
        echo \
'"Content-Type: multipart/mixed; boundary=\"//\"\n'\
'MIME-Version: 1.0\n'\
'\n'\
'--//\n'\
'Content-Type: text/cloud-config; charset=\"us-ascii\"\n'\
'MIME-Version: 1.0\n'\
'Content-Transfer-Encoding: 7bit\n'\
'Content-Disposition: attachment; filename=\"cloud-config.txt\"\n'\
'\n'\
'#cloud-config\n'\
'ssh_pwauth: true\n'\
'disable_root: false\n'\
'chpasswd:\n'\
'  expire: false\n'\
'  users:\n'\
'    - name: root\n'\
'      password: '${root_passwd}'\n'\
'  list: |\n'\
'    root:'${root_passwd}'\n'\
'\n'\
'write_files:\n'\
'  - path: /etc/ssh/sshd_config.d/allow_root.conf\n'\
'    content: |\n'\
'      PermitRootLogin yes\n'\
'      PasswordAuthentication yes\n'\
'\n--//--"'
    ) > $latest_dir/vendor_data.json
fi

dns=$(jq -r .dns <<< $vm_meta)
local_ip=$(jq -r .vlans[0].ip_address <<< $vm_meta)
[ -z "$dns" -o "$dns" = "$local_ip" ] && dns=$dns_server
net_json=$(jq 'del(.userdata) | del(.vlans) | del(.keys) | del(.security) | del(.zvm) | del(.ocp) | del(.virt_type) | del(.dns)' <<< $vm_meta | jq --arg dns $dns '.services[0].type = "dns" | .services[0].address |= .+$dns')
echo "$net_json" > $latest_dir/network_data.json

mkisofs -quiet -R -J -V config-2 -o ${cache_dir}/meta/${vm_ID}.iso $working_dir &> /dev/null
