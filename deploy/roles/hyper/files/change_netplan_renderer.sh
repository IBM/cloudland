#!/bin/bash

wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/local/bin/yq
chmod +x /usr/local/bin/yq
/usr/local/bin/yq '.network.renderer = "NetworkManager"' -i /etc/netplan/50-cloud-init.yaml 
