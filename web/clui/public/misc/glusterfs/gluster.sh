#!/bin/bash

cd $(dirname $0)

[ $# -lt 1 ] && echo "$0 <gluster_id> [heketi_endpoint]" && exit 1
gluster_id=$1
heketi_endpoint=$2
[ -z "$endpoint" ] && heketi_endpoint=http://192.168.91.199:8080

setenforce Permissive
sed -i "s/^SELINUX=enforcing/SELINUX=permissive/" /etc/selinux/config
cp -f /home/centos/.ssh/authorized_keys ~/.ssh/
yum -y install epel-release centos-release-gluster
yum install -y glusterfs-server heketi-client
systemctl start glusterd
systemctl enable glusterd

echo 192.168.91.199 g${gluster_id}-heketi >>/etc/hosts
for i in {0..50}; do
    let suffix=200+$i
    cat >>/etc/hosts <<EOF
192.168.91.$suffix g${gluster_id}-gluster-$i gluster-$i
EOF
done

export HEKETI_CLI_SERVER=$heketi_endpoint
local_ip=$(ifconfig eth0 | grep 'inet ' | awk '{print $2}')
while true; do
    cluster_id=$(heketi-cli cluster list | tail -1 | cut -d: -f2 | cut -d' ' -f1)
    [ -n "$cluster_id" ] && break
    sleep 5
done
while true; do
    node_id=$(heketi-cli node add --zone=1 --cluster=$cluster_id --management-host-name $local_ip --storage-host-name $local_ip | grep 'Id: ' | awk '{print $2}')
    [ -n "$node_id" ] && break
    sleep 5
done
heketi-cli device add --name=/dev/vdb --node=$node_id

host_id=$(hostname -s | cut -d'-' -f3)
let peer_id=${host_id}+1
while true; do
    gluster peer probe g${gluster_id}-gluster-$peer_id
    [ $? -eq 0 ] && break
    sleep 5
done

