#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 1 ] && die "$0 <vm_ID>"

ID=$1
vm_ID=inst-$1
vnc_xml="$vm_ID"_vnc.xml
vnc_pass=`date +%s | sha256sum | base64 | head -c 16`
cp $template_dir/vnc_template.xml $xml_dir/$vm_ID/$vnc_xml
sed -i "s/VNC_PASS/$vnc_pass/g;" $xml_dir/$vm_ID/$vnc_xml
virsh update-device $vm_ID $xml_dir/$vm_ID/$vnc_xml --live
virsh update-device $vm_ID $xml_dir/$vm_ID/$vnc_xml --config
tmpxml=/tmp/${vm_ID}.xml
virsh dumpxml $vm_ID >$tmpxml
vnc_port=$(xmllint --xpath 'string(/domain/devices/graphics/@port)' $tmpxml)
rm -f $tmpxml
local_ip=$(ifconfig $vnc_interface | grep 'inet ' | awk '{print $2}')
iptables -D INPUT -p tcp -m state --state NEW -m tcp -i $vnc_interface --dport $vnc_port -j ACCEPT
iptables -I INPUT -p tcp -m state --state NEW -m tcp -i $vnc_interface --dport $vnc_port -j ACCEPT
echo "iptables -D INPUT -p tcp -m state --state NEW -m tcp -i $vnc_interface --dport $vnc_port -j ACCEPT" | at now + 30 minutes
echo "|:-COMMAND-:| $(basename $0) '$ID' '$vnc_port' '$vnc_pass' '$local_ip'"
