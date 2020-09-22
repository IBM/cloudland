# Adapt the network
To plan the network, you need to know how many IPs you have and how much you can control the devices like hypervisors, switches or routers. Cloudland can make a usable cloud based on your different maneuverability.

## Hypervisor(s) and predefined external IP range for VMs are in the same network

![multi-node-route](images/mnroute.svg)   

   The predefined IPs here mean real route-able IPs and they are accessible as same as the hypervisor IPs. If they belong to different VLANs, refer to the next section. To use this kind of topology, you need to    
1. Delete the existing public subnet from the web UI   
2. Delete device v-5000 from each hypervior
```
nmcli connection delete v-5000
```
3. On each hypervisor, make sure /etc/sysconfig/network-scripts/ifcfg-eth0 and /etc/sysconfig/network-scripts/ifcfg-br5000 look like this:
```
# cat /etc/sysconfig/network-scripts/ifcfg-eth0
TYPE=Ethernet
...
NAME=eth0
DEVICE=eth0
ONBOOT=yes
BRIDGE=br5000
...

# cat /etc/sysconfig/network-scripts/ifcfg-br5000
...
STP=yes
TYPE=Bridge
IPADDR=**node_ip_taken_from_eth0**
PREFIX=**prefix_taken_from_eth0**
GATEWAY=**gateway_taken_from_eth0**
DNS1=**dns_taken_from_eth0**
DEFROUTE=yes
NAME=br5000
DEVICE=br5000
ONBOOT=yes
...
```   
eth0 can be bond0 or what ever the real accessible network device. This config means br5000 takes over eth0's original IP and route table.     
4. Reboot the node and check if the network configuration is all correct   
5. Create new public subnet from web UI
   With your prepared routable subnet/IPs, input the right subnet, netmask, valid IP range like start and end, vlan number 5000, type public.   
6. Now you can try to create a new instance with primary interface of public subnet and test if it is really route-able. 
   Note you may think this is a bit hacking, so it is more desirable if you have more control over the devices like next section.

## You are the admin of hypervisors, IPs and VLANs
    
![real-routing](images/realroute.svg)   
   This means you can control the routing table and IPs. So assume you have hypervisors' eth0 configured on native vlan 120 (192.168.20.0/24), vlan 100 (169.61.25.33/24) is configured for public and vlan 110 (172.16.20.0/24) is configured for private, this is a sample of cisco switch configuration:
```bash
interface GigabitEthernet0/1
 switchport trunk encapsulation dot1q
 switchport trunk native vlan 120
 switchport trunk allowed vlan 100,110,120
 switchport mode trunk
!
interface GigabitEthernet0/2
 switchport trunk encapsulation dot1q
 switchport trunk native vlan 120
 switchport trunk allowed vlan 100,110,120
 switchport mode trunk
!
interface GigabitEthernet0/3
 switchport trunk encapsulation dot1q
 switchport trunk native vlan 120
 switchport trunk allowed vlan 100,110,120
 switchport mode trunk
!
...
interface Vlan100
 ip address 169.61.25.1 255.255.255.0
!
interface Vlan110
 ip address 172.16.20.1 255.255.255.0
!
interface Vlan120
 ip address 192.168.20.1 255.255.255.0
!
...
```
Note with this kind of topology, you probably do not need public access to your hypervisors, so the public IPs in public subnets are planed for instances, the private network is for the instances to access each other across tenants but no need public access.

Now from gui, you can delete both public and private subnets and create new ones accordingly. For public subnet, name public, vlan 100, the ip range aligned with the network device 169.61.25.0/24 here for example; and for private subnet, name private, vlan 110, ip range 172.16.20.0/24

# Deployment
Refer to [deployment guide.md](Deployment)

# Update Configurations 
