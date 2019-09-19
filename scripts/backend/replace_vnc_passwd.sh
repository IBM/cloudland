#!/bin/bash
vm_ID=inst-$1
vnc_xml="$vm_ID"_vnc.xml
vnc_pass=`date +%s | sha256sum | base64 | head -c 16`
echo $vnc_pass
cp $template_dir/vnc_template.xml $xml_dir/$vm_ID/$vnc_xml
sed -i "s/VNC_PASS/$vnc_pass/g;" $xml_dir/$vm_ID/$vnc_xml
virsh update-device $vm_ID $xml_dir/$vm_ID/$vnc_xml --config --live
rm -f $xml_dir/$vm_ID/$vnc_xml
