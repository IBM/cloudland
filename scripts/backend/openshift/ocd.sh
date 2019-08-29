#!/bin/bash

cd $(dirname $0)

[ $# -lt 4 ] && echo "$0 <cluster_name> <base_domain> <external_address> <secret> [ha_flag]"

cluster_name=$1
base_domain=$2
ext_addr=$3
secret=$4
haflag=$5

./oc-dns.sh $cluster_name $base_domain $ext_addr
./oc-lb.sh $cluster_name $base_domain
./oc-ngx.sh

cd /opt
wget https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-client-linux-4.1.11.tar.gz
wget https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-install-linux-4.1.11.tar.gz
tar -zxf *.gz
