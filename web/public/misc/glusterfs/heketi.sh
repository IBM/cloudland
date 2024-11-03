#!/bin/bash

cd $(dirname $0)

[ $# -lt 5 ] && echo "$0 <gluster_id> <endpoint> <cookie> <subnet_id> <nworkers>" && exit 1

gluster_id=$1
endpoint=$2
cookie=$3
subnet_id=$4
nworkers=$5

setenforce Permissive
sed -i "s/^SELINUX=enforcing/SELINUX=permissive/" /etc/selinux/config
yum -y install epel-release centos-release-gluster
yum install -y heketi-client heketi jq
ssh-keygen -f /etc/heketi/heketi_key -t rsa -N ''
key_id=$(curl -k -XPOST -H "X-Json-Format: yes" $endpoint/keys/new --cookie "$cookie" --data "name=heketi_g$gluster_id" --data-urlencode "pubkey=$(cat /etc/heketi/heketi_key.pub)" | jq .ID)
curl -k -XPOST -H "X-Json-Format: yes" $endpoint/glusterfs/$gluster_id --cookie "$cookie" --data "nworkers=$nworkers" --data "heketikey=$key_id"

cat >/etc/heketi/heketi.json <<EOF
{
  "_port_comment": "Heketi Server Port Number",
  "port": "8080",

  "_use_auth": "Enable JWT authorization. Please enable for deployment",
  "use_auth": false,

  "_jwt": "Private keys for access",
  "jwt": {
    "_admin": "Admin has access to all APIs",
    "admin": {
      "key": "My Secret"
    },
    "_user": "User only has access to /volumes endpoint",
    "user": {
      "key": "My Secret"
    }
  },

  "_glusterfs_comment": "GlusterFS Configuration",
  "glusterfs": {
    "_executor_comment": [
      "Execute plugin. Possible choices: mock, ssh",
      "mock: This setting is used for testing and development.",
      "      It will not send commands to any node.",
      "ssh:  This setting will notify Heketi to ssh to the nodes.",
      "      It will need the values in sshexec to be configured.",
      "kubernetes: Communicate with GlusterFS containers over",
      "            Kubernetes exec api."
    ],
    "executor": "ssh",

    "_sshexec_comment": "SSH username and private key file information",
    "sshexec": {
      "keyfile": "/etc/heketi/heketi_key",
      "user": "root",
      "port": "22",
      "fstab": "/etc/fstab"
    },

    "_kubeexec_comment": "Kubernetes configuration",
    "kubeexec": {
      "host" :"https://openshift_api_endpoint:6443",
      "cert" : "/path/to/crt.file",
      "insecure": true,
      "user": "admin_name",
      "password": "admin_password",
      "namespace": "glusterfs_storage",
      "fstab": "/etc/fstab"
    },

    "_db_comment": "Database file name",
    "db": "/var/lib/heketi/heketi.db",

    "_loglevel_comment": [
      "Set log level. Choices are:",
      "  none, critical, error, warning, info, debug",
      "Default is warning"
    ],
    "loglevel" : "debug"
  }
}
EOF

chown -R heketi.heketi /etc/heketi 
echo 192.168.91.199 g${gluster_id}-heketi >>/etc/hosts
for i in {0..50}; do
    let suffix=200+$i
    cat >>/etc/hosts <<EOF
192.168.91.$suffix g${gluster_id}-gluster-$i gluster-$i
EOF
done

systemctl enable heketi
systemctl start heketi
export HEKETI_CLI_SERVER=http://192.168.91.199:8080
while true; do
    heketi-cli cluster create
    [ $? -eq 0 ] && break
    sleep 5
done
