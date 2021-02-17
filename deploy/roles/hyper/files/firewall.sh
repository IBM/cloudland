#!/bin/bash

console_ip=$1

if [ $# -eq 2 ]; then
    bridge="br"$2
fi

iptables -D INPUT -m state --state RELATED,ESTABLISHED -j ACCEPT
iptables -D INPUT -p icmp -j ACCEPT
iptables -D INPUT -i lo -j ACCEPT
iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 22 -j ACCEPT
iptables -D INPUT -p tcp -m tcp --dport 24007:24009 -m conntrack --ctstate NEW -j ACCEPT
iptables -D INPUT -p tcp -m tcp --dport 6188 -m conntrack --ctstate NEW -j ACCEPT
iptables -D INPUT -p tcp -m tcp --dport 49152:49664 -m conntrack --ctstate NEW -j ACCEPT
iptables -D INPUT -p tcp -m tcp --dport 38465:38469 -m conntrack --ctstate NEW -j ACCEPT
iptables -D INPUT -p udp -m udp --dport 8472 -m conntrack --ctstate NEW -j ACCEPT
iptables -D INPUT -p tcp -s $console_ip -m conntrack --ctstate NEW -j ACCEPT
iptables -D INPUT -j REJECT --reject-with icmp-host-prohibited
iptables -D FORWARD -j REJECT --reject-with icmp-host-prohibited
if [ $# -eq 2 ]; then
    iptables -D FORWARD -i $bridge -o $bridge -j ACCEPT
fi
iptables -A INPUT -m state --state RELATED,ESTABLISHED -j ACCEPT
iptables -A INPUT -p icmp -j ACCEPT
iptables -A INPUT -i lo -j ACCEPT
iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 22 -j ACCEPT
iptables -A INPUT -p tcp -m tcp --dport 24007:24009 -m conntrack --ctstate NEW -j ACCEPT
iptables -A INPUT -p tcp -m tcp --dport 6188 -m conntrack --ctstate NEW -j ACCEPT
iptables -A INPUT -p tcp -m tcp --dport 49152:49664 -m conntrack --ctstate NEW -j ACCEPT
iptables -A INPUT -p tcp -m tcp --dport 38465:38469 -m conntrack --ctstate NEW -j ACCEPT
iptables -A INPUT -p udp -m udp --dport 8472 -m conntrack --ctstate NEW -j ACCEPT
iptables -A INPUT -p tcp -s $console_ip -m conntrack --ctstate NEW -j ACCEPT
iptables -A INPUT -j REJECT --reject-with icmp-host-prohibited
if [ $# -eq 2 ]; then
    iptables -I FORWARD -i $bridge -o $bridge -j ACCEPT
fi
iptables -A FORWARD -j REJECT --reject-with icmp-host-prohibited

iptables -P FORWARD DROP
iptables -N secgroup-chain && iptables -A secgroup-chain -j ACCEPT

service iptables save
