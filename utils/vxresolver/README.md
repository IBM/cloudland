[![Build Status](https://travis.ibm.com/cloudland/vxresolver.svg?token=V922FVR3sGqicTaYbfDJ&branch=master)](https://travis.ibm.com/cloudland/vxresolver)

# vxresolver

## Create a vxlan device

   ```
# Replace 5000 with any number, use 4789 here so it can co-work with OVS    
ip link add vxlan-5000 type vxlan id 5000 dev eth0 dstport 4789   

# Set target for BUM packets
bridge fdb add ff:ff:ff:ff:ff:ff dev vxlan-5000 dst 169.254.169.254 self permanent 
   ```

## Configure iptables
   ```
iptables -t nat -A OUTPUT -d 169.254.169.254/32 -p udp -m udp --dport 4789 -j DNAT --to-destination 127.0.0.1:8896
   ```

## Dtabase schema
   ```
CREATE TABLE VTEP (id INTEGER PRIMARY KEY AUTOINCREMENT, instance varchar(32), vni INTEGER, inner_ip varchar(32), inner_mac varchar(48), outer_ip varchar(32));
   ```
## Start vxresolver
   ```
./vxresolver -d /opt/cloudland/db/resolve.db -x 4789
   ```
