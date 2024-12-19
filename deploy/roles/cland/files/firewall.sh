#!/bin/bash

iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 5006 -j ACCEPT
iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 5006 -j ACCEPT

/sbin/iptables-save -c > /etc/iptables.rules
