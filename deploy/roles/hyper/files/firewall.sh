#!/bin/bash

iptables -D FORWARD -j REJECT --reject-with icmp-host-prohibited
iptables -A FORWARD -j REJECT --reject-with icmp-host-prohibited

iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 6188 -j ACCEPT
iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 6188 -j ACCEPT

for chain in $(iptables -S | grep secgroup | awk '{print $2}'); do
    iptables -X $chain
done

/sbin/iptables-save -c > /etc/iptables.rules
rm -f /opt/cloudland/run/need_to_sync
