#!/bin/bash

YQ=/tmp/yq
wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O $YQ
chmod +x $YQ
renderer=$($YQ '.network.renderer' /etc/netplan/*.yaml)
if [ "$renderer" != "NetworkManager" ]; then
    $YQ '.network.renderer = "NetworkManager"' -i /etc/netplan/*.yaml
    echo need_to_reboot
    exit 0
fi
rm -f $YQ
echo no_need_to_reboot
