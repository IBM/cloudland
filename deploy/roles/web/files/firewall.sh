#!/bin/bash

# clbase
iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 5005 -j ACCEPT
iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 5005 -j ACCEPT
iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 5443 -j ACCEPT
iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 5443 -j ACCEPT

# clapi
iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 8255 -j ACCEPT
iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 8255 -j ACCEPT

# virtconsoleproxyd
iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 9443 -j ACCEPT
iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 9443 -j ACCEPT

/sbin/iptables-save -c > /etc/iptables.rules
