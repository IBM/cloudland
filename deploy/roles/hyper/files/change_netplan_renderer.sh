#!/bin/bash

if [ ! -x /usr/local/bin/yq ]; then
    wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/local/bin/yq >/dev/null 2>&1
    chmod +x /usr/local/bin/yq
fi
renderer=$(/usr/local/bin/yq '.network.renderer' /etc/netplan/50-cloud-init.yaml)
if [ "$renderer" != "NetworkManager" ]; then
    /usr/local/bin/yq '.network.renderer = "NetworkManager"' -i /etc/netplan/*.yaml
    echo need_to_reboot
fi
