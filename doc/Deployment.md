# Development Deployment
For development, you can prepare one or more bare metals or VMs with nested kvm enabled to have a fast installation. 

## Suggested hardware specs
* CPU: >=4
* Memory: >=8G
* Disk: disk1 >= 500G, disk2 >= 500G, disk1 is for compute and disk2 is for storage. For development or a quick trying, it is fine to use a single small disk like 10G. 

## Prerequisite for the bare metals (or VMs)
* CentOS 7.5+   
* Create user cland with sudoer privilege 
```bash
    useradd cland
    echo 'cland ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers.d/cland
```
* Make sure hostname not 'localhost'
 
## Install all in one   
* With cland user create above for all the commands
```bash
    sudo yum -y install git
    cd /opt
    sudo git clone https://github.com/IBM/cloudland.git  # Suggest to use your own forked repository
    cd /opt/cloudland/deploy
```
* Create netconf.yml from netconf.yml.example and modify the variables according to your environment
* Make sure the interfaces in netconf.xml are controlled by NetworkManager
* Execute allinone.sh, this step takes about 20 minutes
```bash
    ./allinone.sh
```   
## Verify the installation
* Open browser with url ```http://<allinone_ip_or_dns>```
* Click flavors, images, subnets at the left side and check the items in the right side. There should be items of flavor m1.tiny, image cirros.qcow2 and subnets public and private.
* The public and private subnets are public external and private external, this is just for demo and development purpose. To make cloudland provisioned VMs actually routable, please refer to network section of the [operation guide](Operation).
* Launch an instance from image cirros
* For more usage, refer to [user manual](Manual)
 
## Add more compute nodes
* Prepare ansible hosts file    
   
   On allinone node, Edit server name, ansible_host ... of hyper section in ```/opt/cloudland/deploy/hosts/hosts``` file to make like below; Note client_id is the backend id assigned to each compute nodes; It must start at 0 and increase by 1
```bash
    [hyper]
    server-0 ansible_host=192.168.10.20 ansible_ssh_private_key_file=/opt/cloudland/deploy/.ssh/cland.key client_id=0
    server-1 ansible_host=192.168.10.13 ansible_ssh_private_key_file=/opt/cloudland/deploy/.ssh/cland.key client_id=1
```
* Add cland user and pub key to the new compute nodes   
   
   After the cland user creation described in prerequisite on the new compute nodes, append ```/opt/cloudland/deploy/.ssh/cland.key.pub``` from allinone to the new nodes' ```/home/cland/.ssh/authorized_keys``` file to make sure the following deployment without password prompts   

* Install glusterfs on all compute nodes      
* Run the deployment
```bash
    cd /opt/cloudland/deploy
    ansible-playbook cloudland.yml --tags hosts,epel,ntp,fe_srv,hyper --skip-tags be_conf
```

* Check the new nodes   
   
   Check the connections between frontend and backends -- cloudland and cloudlets, if all hypers are connected
```bash
    sudo netstat -nap | grep cloudland
```   

## Adjust Networks
Refer to network section in [operation guide](Operation)
# Production Deployment   
Production environment can be upgraded from development environment. With a complete working development environment, you can create a few VMs to hold the clustered control plain, migrate the data to the new control plain.