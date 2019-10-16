#!/bin/bash

setenforce Permissive
sed -i "s/^SELINUX=enforcing/SELINUX=permissive/" /etc/selinux/config
cp -f /home/centos/.ssh/authorized_keys ~/.ssh/
yum install -y glusterfs-server heketi-client
systemctl start glusterd
systemctl enable glusterd

echo 192.168.91.199 g${gluster_id}-heketi >>/etc/hosts
for i in {0..50}; do
    let suffix=200+$i
    cat >>/etc/hosts <<EOF
192.168.91.$suffix g${gluster_id}-gluster-$i
EOF
done

host_id=$(hostname -s | cut -d'-' -f3)
if [ "$host_id" -gt 0 ]; then
    let peer_id=$host_id-1
    while true; do
        gluster peer probe g${gluster_id}-$peer_id
        [ $? -eq 0 ] && break
    done
fi

export HEKETI_CLI_SERVER=http://192.168.91.199:8080
local_ip=$(ifconfig eth0 | grep 'inet ' | awk '{print $2}')
cluster_id=$(heketi-cli cluster list | tail -1 | cut -d: -f2 | cut -d' ' -f1)
node_id=$(heketi-cli node add --zone=1 --cluster=$cluster_id --management-host-name $local_ip --storage-host-name $local_ip | cut -d: -f2 | cut -d' ' -f1)
#heketi-cli device add --name=/dev/vdb --node=$node_id

