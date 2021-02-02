# Steps to install CloudLand
When all the [prequisites](#prequisites) are met and the [conf.json](#confjson) is prepared:
1. Install from RPM package (from controller):
    ```
   # Switch to cland
   su - cland

   # Install via yum
   yum -y install cloudland-<ver>-<rel>.<arch>.rpm

   # Copy conf.json to /opt/cloudland/deploy
   cp <path-to>/conf.json /opt/cloudland/deploy

   # Deploy controller and compute nodes
   cd /opt/cloudland/deploy
   ./deploy.sh

   # Verify: access https://[controller-ip] from web browser and use default "admin/passw0rd" to login. Please change admin's password immediately.

2. Install from source code (from controller, which is also the build server):
    ```
    # Build from source
    # ...

    # Switch to cland
    su - cland

    # Copy conf.json to /opt/cloudland/deploy
    cp <path-to>/conf.json /opt/cloudland/deploy

    # Deploy controller and compute nodes
    cd /opt/cloudland/deploy
    ./deploy.sh

    # Verify: access https://[controller-ip] from web browser and use default "admin/passw0rd" to login. Please change admin's password immediately.
    ```
3. Deploy new compute node(s) after installation
   ```
   # Prepare new compute node
   # ...

   # Add new compute node(s) configuration to conf.json
   # Important: all ids must be continuous in conf.json
   # ...

   # Deploy new compute node, inclusive [begin_id, end_id]
   cd /opt/cloudland/deploy
   ./deploy_compute.sh begin_id end_id
   ```
4. Update compute node(s) after installation
    ```
    # Prepare new compute node if the configurations are changed
    # ...

    # Update the configurations of conf.json if they are changed
    # ...

    # Build and install new binaries if source codes are changed
    # ...

    # Update compute node, inclusive [begin_id, end_id]
    cd /opt/cloudland/deploy
    ./deploy_compute.sh begin_id end_id

# About build server, controller and compute nodes
Logically, there are three types of roles in CloudLand:
1. Build server: 
   - refer to [Build](Build.md) for more information. It's used to build the binaries.
2. Controller:
   - The entry point which users use to manange all the resources, like compute node (hypervisors), network, VMs, images, etc.
   - User accesses controller via https://[controller-ip]
3. Compute nodes
   - The nodes which create VMs.

Note:
1. In current release, the architecture of all nodes are the same, like s390x, or x86-64, etc.
2. For development, the Build Server and Controller can be the same machine. After [building](Build.md), CloudLand can be installed directly after [preparing compute nodes](#prequisites) and the [conf.json](#confjson)
3. Controller can be a compute node too, which means the deploy.sh will apply the compute node role to controller.

# [Prequisites](#prequisites)
## For Controller:
1. yum is installed and it can install following softwares: 
   - **ansible wget jq net-tools gnutls-utils iptables iptables-services postgresql postgresql-server postgresql-contrib**
   - Note: the deploy.sh will try to use yum to install the softwares, if you want to install them yourselves without yum, please comment out the "yum install" related code from deploy.sh
2. user '**cland**' and group '**cland**' are added
   ```
   useradd cland
   echo 'cland ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers.d/cland
   groupadd cland
   ```
3. **cland.key** and **cland.key.pub** are generated under **/home/cland/.ssh** and cland.key.pub has been added to /home/cland/.ssh/authorized_keys (CloudLand uses ansible to deploy compute node, it's possible that the controller itself is a compute node, so we need the public key installed)
   ```
   # create .ssh and authorized_keys if they don't exist, and change their mode
   mkdir -p ~/.ssh
   chmod 700 ~/.ssh
   touch ~/.ssh/authorized_keys
   chmod 600 ~/.ssh/authorized_keys

   # generate cland.key and cland.key.pub
   yes y | ssh-keygen -t rsa -N "" -f /home/cland/.ssh/cland.key

   # add cland.key.pub to authorized_keys
   cat /home/cland/.ssh/cland.key.pub >> /home/cland/.ssh/authorized_keys
   ```

## For each compute nodes:
1. yum is installed and it can install following softwares:
   1. All needs: **sqlite jq mkisofs NetworkManager net-tools iptables-services**
   2. For KVM and KVM on Z: 
      1. **kvm virtualization is enabled** 
      2. **qemu-img qemu-img libvirt libvirt-client dnsmasq keepalived dnsmasq-utils conntrack-tools**
   3. For z/VM:
      1. (Current release) Assume [feilong](https://github.com/openmainframeproject/feilong) has been installed and it's providing service via http://127.0.0.1:8080, refer to its [repo](https://github.com/openmainframeproject/feilong) and [document](https://cloudlib4zvm.readthedocs.io/en/latest/index.html) for more information.
   4. **Note**: when deploying the compute node, ansible will try to install the software via yum but the error will be ignored. Please check the ansible log and the compute node to make sure these softwares are installed, otherwise, the compute node will not work.
2. user '**cland**' and group '**cland**' are added
   ```
   useradd cland
   echo 'cland ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers.d/cland
   groupadd cland
   ```
3. The **cland.key.pub** from controller are added to the **/home/cland/.ssh/authorized_keys** on each compute node
    - use ssh-copy-id from controller
    - or, copy cland.key.pub from controller and paste the content to the /home/cland/.ssh/authorized_keys on each compute node directly
4. (Current release) Network requirement for KVM and KVM on Z:
   1. Refer to [Operation](Operation.md) to configure the network manually. 
   2. When configure the [conf.json](#confjson), set the "network_external_vlan" and "network_internal_vlan" according to the network configuration, like bridge 5000 and 5010.

# [conf.json](#confjson)
- conf.json is the configuration file which describes the controller and compute nodes. There's an example: /opt/cloudland/deploy/conf.json.example . Copy the json part to conf.json and update it according to the real configuratons of the controller and each compute nodes.
- The conf.json is used to generate the cloudrc.local file which will be used by the compute node when doing the real jobs. After deployment, check /opt/cloudland/scripts/cloudrc.local for more information

## Important:
1. The controller IP is the entry point to access CloudLand after installation.
2. The sequence of the compute node id is mandatory: it start from 0, increases by 1. 
3. There are three virt_types: **zvm, zkvm and kvm**:
   1. **zvm** is for z/VM hypervisor. In current release, this kind of hypervisor relies on [felong](https://github.com/openmainframeproject/feilong). It should be installed and be providing service from http://127.0.0.1:8080 (the default service point) on the compute node. The default guest name is ZCCXXXXX, you can find them from the cloudrc.local, which are not included in conf.json
   2. **zkvm**, the KVM on Z, it's the KVM hypervisor running on Z. The settings are the same as the KVM, but it has one more entry called 'zlayer2_iface' which is used to configure the fdb entries.
   3. **kvm**, the KVM on non-s390x, like x86-64.
   4. Note: In current release, we only support the same architecture node, which means CloudLand can support zvm and zkvm at the same time(s390x for all build server, controller and compute nodes), or kvm (x86-64 build server controller and compute nodes) only.
4. The zone_name should be pre-set according to the whole topology. 