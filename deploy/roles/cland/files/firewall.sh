iptables -D INPUT -p tcp -m tcp --dport 9988 -m conntrack --ctstate NEW -j ACCEPT
iptables -D INPUT -p tcp -m tcp --dport 50051 -m conntrack --ctstate NEW -j ACCEPT
iptables -D INPUT -p tcp -m tcp --dport 9443 -m conntrack --ctstate NEW -j ACCEPT
iptables -I INPUT -p tcp -m tcp --dport 9988 -m conntrack --ctstate NEW -j ACCEPT
iptables -I INPUT -p tcp -m tcp --dport 50051 -m conntrack --ctstate NEW -j ACCEPT
iptables -I INPUT -p tcp -m tcp --dport 9443 -m conntrack --ctstate NEW -j ACCEPT

iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 80 -j ACCEPT
iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 80 -j ACCEPT
iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 443 -j ACCEPT
iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 443 -j ACCEPT
iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 4000 -j ACCEPT
iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 4000 -j ACCEPT
iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 9988 -j ACCEPT
iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 9988 -j ACCEPT
iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 18000:20000 -j ACCEPT
iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 18000:20000 -j ACCEPT
service iptables save