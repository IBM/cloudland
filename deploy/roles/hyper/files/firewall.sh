#!/bin/bash

iptables -D FORWARD -j REJECT --reject-with icmp-host-prohibited
iptables -A FORWARD -j REJECT --reject-with icmp-host-prohibited

iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 6188 -j ACCEPT
iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport 6188 -j ACCEPT

/sbin/iptables-save -c > /etc/iptables.rules

cat >/etc/NetworkManager/dispatcher.d/01firewall <<EOF
if [ -x /usr/bin/logger ]; then
        LOGGER="/usr/bin/logger -s -p daemon.info -t FirewallHandler"
else
        LOGGER=echo
fi

case "\$2" in
        up)
                if [ ! -r /etc/iptables.rules ]; then
                        \${LOGGER} "No iptables rules exist to restore."
                        return
                fi
                if [ ! -x /sbin/iptables-restore ]; then
                        \${LOGGER} "No program exists to restore iptables rules."
                        return
                fi
                \${LOGGER} "Restoring iptables rules"
                /sbin/iptables-restore -c < /etc/iptables.rules
                ;;
        down)
                if [ ! -x /sbin/iptables-save ]; then
                        \${LOGGER} "No program exists to save iptables rules."
                        return
                fi
                \${LOGGER} "Saving iptables rules."
                /sbin/iptables-save -c > /etc/iptables.rules
                ;;
        *)
                ;;
esac
EOF
chmod +x /etc/NetworkManager/dispatcher.d/01firewall
