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

vm_meta=$(cat | base64 -d)
userdata=$(jq -r .userdata <<<$vm_meta)
if [ -n "$userdata" ]; then
   echo "$userdata" > $latest_dir/user_data
fi

root_passwd=$(jq -r '.root_passwd' <<< $vm_meta)
os_code=$(jq -r '.os_code' <<< $vm_meta)
dns=$(jq -r '.dns' <<< $vm_meta)
login_port=$(jq -r '.login_port' <<< $vm_meta)
pub_keys=$(jq -r '.keys' <<< $vm_meta)
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
    if [ -n "${root_passwd}" ] && [ "${os_code}" = "windows" ]; then
        echo '  "admin_pass": "'${root_passwd}'",'
    else
        echo '  "admin_pass": "'${admin_pass}'",'
    fi
    echo '  "random_seed": "'${random_seed}'"'
    echo '}'
) > $latest_dir/meta_data.json

ssh_pwauth="false"
if [ -n "${root_passwd}" ] && [ "${os_code}" != "windows" ]; then
    ssh_pwauth="true"
fi

# vendor_data.json header
vendor_data_header=$( 
    echo \
'"Content-Type: multipart/mixed; boundary=\"//\"\n'\
'MIME-Version: 1.0\n'\
'\n'\
'--//\n'
)

# cloud-config.txt header
cloud_config_txt=$(
    echo \
'Content-Type: text/cloud-config; charset=\"us-ascii\"\n'\
'MIME-Version: 1.0\n'\
'Content-Transfer-Encoding: 7bit\n'\
'Content-Disposition: attachment; filename=\"cloud-config.txt\"\n'\
'\n'\
'#cloud-config\n'\
'ssh_pwauth: '${ssh_pwauth}'\n'\
'disable_root: false\n'\
)

# cloud-config.txt body
if [ -n "${root_passwd}" ] && [ "${os_code}" != "windows" ]; then
    cloud_config_txt+=$(
        echo \
'chpasswd:\n'\
'  expire: false\n'\
'  users:\n'\
'    - name: root\n'\
'      password: '${root_passwd}'\n'\
'  list: |\n'\
'    root:'${root_passwd}'\n'\
'write_files:\n'
'  - path: /etc/ssh/sshd_config.d/allow_root.conf\n'\
'    content: |\n'\
'      PermitRootLogin yes\n'\
'      PasswordAuthentication yes\n'
    )
fi

# use runcmd to change the port value of /etc/ssh/sshd_config
# and restart the ssh service
if [ -n "${login_port}" ] && [ "${login_port}" != "22" ] && [ "${os_code}" != "windows" ]; then
    cloud_config_txt+=$(
        echo \
'runcmd:\n'\
'    - sed -i \"s/^#Port .*/Port '${login_port}'/\" /etc/ssh/sshd_config\n'\
'    - sed -i \"s/^Port .*/Port '${login_port}'/\" /etc/ssh/sshd_config\n'\
'    - systemctl restart sshd\n'
    )
fi
    
vendor_data_end='\n--//--"'

# write to vendor_data.json
if [ "${os_code}" != "windows" ]; then
    echo -e "$vendor_data_header""$cloud_config_txt""$vendor_data_end" > $latest_dir/vendor_data.json
    sed -i -n '1h; 1!H; ${ x; s/\n/\\n/g; p; }' $latest_dir/vendor_data.json
fi

# write script part windows powershell script to change the RDP port 
# and restart the RDP service
# and configure windows firewall to allow the new RDP port
# if [ -n "${login_port}" ] && [ "${login_port}" != "3389" ] && [ "${os_code}" = "windows" ]; then
#     (
#         echo '<powershell>\r'
#         echo '$new_port = '${login_port}'\r'
#         echo '$old_port = (Get-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\Terminal Server\WinStations\RDP-Tcp\").PortNumber\r'
#         echo 'if ($new_port -ne $old_port) {\r'
#         echo '  Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\Terminal Server\WinStations\RDP-Tcp" -Name PortNumber -Value $new_port\r'
#         echo '  Restart-Service -Name "TermService" -Force\r'
#         echo '  New-NetFirewallRule -DisplayName "RDP-TCP-$($new_port)" -Action Allow -Protocol TCP -LocalPort $new_port\r'
#         echo '}\r'
#         echo '</powershell>'
#     ) > $latest_dir/change_rdp_port.ps1
# fi

[ -z "$dns" ] && dns=$dns_server
net_json=$(jq 'del(.userdata) | del(.vlans) | del(.keys) | del(.security) | del(.login_port) | del(.root_passwd) | del(.dns)' <<< $vm_meta | jq --arg dns $dns '.services[0].type = "dns" | .services[0].address |= .+$dns')
let mtu=$(cat /sys/class/net/$vxlan_interface/mtu)-50
if [ "$mtu" -lt 1450 ]; then
    net_json=$(jq ".links[] | .mtu = $mtu" <<<$net_json)
fi
echo "$net_json" > $latest_dir/network_data.json

mkisofs -quiet -R -J -V config-2 -o ${cache_dir}/meta/${vm_ID}.iso $working_dir &> /dev/null
