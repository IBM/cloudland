iptables -D INPUT -p tcp -m tcp --dport 9988 -m conntrack --ctstate NEW -j ACCEPT
iptables -D INPUT -p tcp -m tcp --dport 50051 -m conntrack --ctstate NEW -j ACCEPT
iptables -I INPUT -p tcp -m tcp --dport 9988 -m conntrack --ctstate NEW -j ACCEPT
iptables -I INPUT -p tcp -m tcp --dport 50051 -m conntrack --ctstate NEW -j ACCEPT
