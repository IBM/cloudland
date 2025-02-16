#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 2 ] && die "$0 <vm_ID> <vnc_pass>"

ID=${1##inst-}
vnc_pass=$2
vm_ID=inst-$ID
vnc_xml="$vm_ID"_vnc.xml
cp $template_dir/vnc_template.xml $xml_dir/$vm_ID/$vnc_xml
sed -i "s/VNC_PASS/$vnc_pass/g;" $xml_dir/$vm_ID/$vnc_xml
timeout_virsh update-device $vm_ID $xml_dir/$vm_ID/$vnc_xml --live
timeout_virsh update-device $vm_ID $xml_dir/$vm_ID/$vnc_xml --config
tmpxml=/tmp/${vm_ID}.xml
timeout_virsh dumpxml $vm_ID >$tmpxml
vnc_port=$(xmllint --xpath 'string(/domain/devices/graphics/@port)' $tmpxml)
rm -f $tmpxml
local_ip=$(ifconfig $vxlan_interface | grep 'inet ' | awk '{print $2}')
echo "|:-COMMAND-:| $(basename $0) '$ID' '$vnc_port' '$local_ip'"
