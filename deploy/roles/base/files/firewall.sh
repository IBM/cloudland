#!/bin/bash

iptables -P INPUT ACCEPT
iptables -P FORWARD ACCEPT
iptables -P OUTPUT ACCEPT
iptables -F
iptables -D INPUT -m state --state RELATED,ESTABLISHED -j ACCEPT
iptables -I INPUT -m state --state RELATED,ESTABLISHED -j ACCEPT
iptables -D INPUT -p icmp -j ACCEPT
iptables -A INPUT -p icmp -j ACCEPT
iptables -D INPUT -i lo -j ACCEPT
iptables -A INPUT -i lo -j ACCEPT
iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 22 -j ACCEPT
iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 22 -j ACCEPT
iptables -D INPUT -s 10.0.0.0/8 -j ACCEPT
iptables -I INPUT -s 10.0.0.0/8 -j ACCEPT
iptables -D INPUT -s 172.16.0.0/12 -j ACCEPT
iptables -I INPUT -s 172.16.0.0/12 -j ACCEPT
iptables -D INPUT -s 192.168.0.0/16 -j ACCEPT
iptables -I INPUT -s 192.168.0.0/16 -j ACCEPT
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

/sbin/iptables-save -c > /etc/iptables.rules

cat >/etc/network/if-pre-up.d/iptablesload <<EOF
#!/bin/sh
iptables-restore < /etc/iptables.rules
exit 0
EOF

cat >/etc/network/if-post-down.d/iptablessave <<EOF
#!/bin/sh
iptables-save -c > /etc/iptables.rules
if [ -f /etc/iptables.downrules ]; then
   iptables-restore < /etc/iptables.downrules
fi
exit 0
EOF

chmod +x /etc/network/if-post-down.d/iptablessave
chmod +x /etc/network/if-pre-up.d/iptablesload

cat >/etc/networkd-dispatcher/routable.d/50-ifup-hooks <<EOF
#!/bin/sh

for d in up post-up; do
    hookdir=/etc/network/if-${d}.d
    [ -e $hookdir ] && /bin/run-parts $hookdir
done
exit 0
EOF

cat >/etc/networkd-dispatcher/off.d/50-ifdown-hooks <<EOF
#!/bin/sh

for d in down post-down; do
    hookdir=/etc/network/if-${d}.d
    [ -e $hookdir ] && /bin/run-parts $hookdir
done
exit 0
EOF

chmod +x /etc/networkd-dispatcher/routable.d/50-ifup-hooks
chmod +x /etc/networkd-dispatcher/off.d/50-ifdown-hooks
