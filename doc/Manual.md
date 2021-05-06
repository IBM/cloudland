# Create a key
To create a key, simply input key name and public key into fields and submit. If you don't have a key already, use command ```ssh-keygen -f /path/to/your_key``` to generate one. 

# Create an image
Input image name, url and architecture into create image tab and submit, wait till the image status become available. Centos7 image is created with the environment deployment.

# Create a flavor
With allinone installation, a flavor named m1.tiny is created already, you can create more with your desired number of cpu, memory and disk size. Only admin has the privilege. 

# Create a subnet
You can specify a name with valid network, netmask, gateway, start and end into fields and submit. Make sure no conflict among your fields. Admin can create public or private and specify a vlan number, normal user can only create internal subnet with a random vxlan id. For what is public or private vlan, refer to network section in 
[operation guide](Operation). A default subnet with username is created with user creation.

# Create a security group with rules
A default security group named username is created when user is created. You can modify it, delete it or create new ones. The default security group has icmp and port 22 opened.

# Launch an instance
With all the above, it is straightforward to create an instance via web ui. if you are admin, you can create instance on public or private subnets directly, otherwise you can only create instance on your own internal subnets. To access the instance on internal subnets, you must create a gateway, and then create a floating ip to access it. You can also edit or view the instance by clicking instance ID. To get VNC password, you need to refresh the edit page.

# Create a gateway
To create a gateway, you can specify a name or choose what kind of network the gateway wants to route, it can be public, private or both, if you don't know what it is, leave it blank.

# Create a floating IP
After you create a gateway, you can choose a instance from the drop down menu and create a floating ip to it. 

# Create an OpenShift cluster

## Prerequisite Of Openshift Cluster Platform Registry

   - Generate pull secret according to https://cloud.redhat.com/openshift/install/pull-secret

## Fill in the registry information in Registry tab
	- Label: the registry name
	- Virtualization Type:
		- KVM on Z
		- KVM on x86_64
		- z/VM
	- Openshift Version: version 4.6 and higher are supported
	- Registry Content: 
		  pullSecret: '{"auths": ...}'
		  sshKey: 'ssh-ed25519 AAAA...'
		  <optional, it is required for private registry>
	      additionalTrustBundle: | 
                  -----BEGIN CERTIFICATE-----
                    <MY_TRUSTED_CA_CERT>
                  -----END CERTIFICATE----- 
		  <optional, it is required for private registry>              
 		  imageContentSources:
		  - mirrors:
		    - ...
                  source: ... 
          - mirrors:
		    - ...
                  source: ...
	- RHCOS initramfs: Location of the initramfs file. For example:
		Local: 
		file:///<local_path>/rhcos-<version>-live-initramfs-<architecture>
		Remote (Need a http server): 
		http://<HTTP_server>/<path>/rhcos-<version>-live-initramfs-<architecture> 
	- RHCOS kernel: Localtion of the kernel file. For example
		Local: 
		file:///<local_path>/rhcos-<version>-live-kernel-<architecture>	     
        Remote (Need a http server): 
		http://<HTTP_server>/<path>/rhcos-<version>-live-kernel-<architecture>
	- RHCOS image: Location of the rootfs file location. For example
		Local: 
		file:///<local_path>/rhcos-<version>-live-rootfs-<architecture>	
		Remote (Need a http server): 
		http://<HTTP_server>/<path>/rhcos-<version>-live-rootfs-<architecture>
	- OCP installer: Location of the openshift-install. For example
        Local: 
		file:///<local_path>/openshift-install-linux.tar.gz
  		Remote (Need a http server): 
		http://<HTTP_server>/<path>/openshift-install-linux.tar.gz
	- OCP client: Location for the openshift-client. For example
		Local: 
		file:///<local_path>/openshift-client-linux.tar.gz
        Remote (Need a http server): 
		http://<HTTP_server>/<path>/openshift-client-linux.tar.gz

* CloudLand uses curl to download the files.
* Download the RHCOS images from RedHat offical website and save it to local or a http server before creating the registry.
    
## Click the **Create** button to generate the registry

- Go to the directory (/opt/cloudland/cache/image/ocp/[**version**]/[**virtType**]) to check if the files are downloaded successfully or not.

## Delete or update the registry information in Registry Tab.

## Config your own yum repo for private registry(Optional)

- Prepare your own yum repo file with the filename format of **ocd_lb_yum_${virt_type}.repo** and upload to **/opt/cloudland/web/clui/public/misc/openshift** in the Controller Server.

## Create a Openshift Cluster
	
* Go to **Openshift** tab and click **Create** button:
	- Cluster Name
	- Base Domain: determine the domain names of your access endpoint and APPs
		the combination of `${cluster_name}.${base_domain}` must be a valid domain name.
	- Zone: zone to install the Openshift cluster
	- Infrastructure Type:
		KVM on Z
		KVM on x86_64
		z/VM
	- LoadBalancer_IP: <Optional>.
	- High Available: <Optional> The Default Setting is no.
	- Number of Workers: The Default Setting is 2
	- Registry: Select or Create if it's necessary
	- LB_Flavor: Select or Create if it's necessary
	- Master_Flavor: Select or Create if it's necessary
	- Worker_Flavor: Select or Create if it's necessary
	- Subnet: <Optional> The Default Setting is Public. Choose or Create if it's necessary
	- Key: key to login Load balancer. Choose or Create if it's necessary.

* Check the instances from the **Instances** tab.  
  * The instances named **ocID-lb, bootstrap, master-N, workers-N**  
  * The instance ocID-lb is the Load Balancer and domain name server for the cluster, you can login it via its floating IP with the key specified before and see the log file /tmp/ocd.log.

## Login Openshift in Web Console

Once the cluster status is marked as complete, you can get the credentials from lb:/opt/\${cluster_name}/auth. 

The web console is https://console-openshift-console.apps.\${cluster_name}.\${base_domain}.

There are 3 ways to access the DNS records of the APPs in this cluster
1. Ideally, if you are the owner of `${base_domain}`, then in your DNS provider's website, you can create: 
   * An 'A' record with `dns.${cluster_name}.${base_domain}` pointing to the public floating IP of instance ocID-lb
   * An 'NS' record with `${cluster_name}.${base_domain}` referring to `dns.${cluster_name}.${base_domain}`
2. For temporary usage or testing purpose, you can modify file /etc/hosts in your working machine with these records for example:
	```
	$floating_ip     console-openshift-console.apps.${cluster_name}.${base_domain}
	$floating_ip     oauth-openshift.apps.${cluster_name}.${base_domain}
	$floating_ip     prometheus-k8s-openshift-monitoring.apps.${cluster_name}.${base_domain}
	$floating_ip     grafana-openshift-monitoring.apps.${cluster_name}.${base_domain}
	$floating_ip     alertmanager-main-openshift-monitoring.apps.${cluster_name}.${base_domain}
	$floating_ip     downloads-openshift-console.apps.${cluster_name}.${base_domain}
	$floating_ip     default-route-openshift-image-registry.apps.${cluster_name}.${base_domain}
	```
	If you have more apps created, then create more records accordingly and similarly.

3. Alternatively, you can modify /etc/resolv.conf in your working machine with
	```
	nameserver $floating_ip
	```
	The ocID-lb is able to resolve the domain names both inside and outside of the cluster.

**Note**: 
1. the installation procedure uses your access cookie, so do not logout cloudland web console before the installation completes.

### Scale up/down workers

To scale up/down more/less workers, click the cluster ID and input worker number in edit page and submit   
1. configuration in ocID-lb need to be updated after scale up/down
   * update **/etc/dnsmasq.openshift.addnhosts** to make sure the ip of worker-N.${cluster_name}.${base_domain} is correct.
   * update **/etc/haproxy/haproxy.cfg** to make sure the corresponding port (80,443) of worker-x.${cluster_name}.${base_domain} is correct.
2. restart haproxy and dnsmaq to take effect
   * **systemctl restart haproxy**
   * **systenctl restart dnsmasq**
3. to approve csr certification by manual when scalling up
   * **oc get csr** to review the status of csr certification
   * **oc adm certificate approve xxxx** to approve csr certification. **xxxx** means csr NAME
4. delete worker node by manual when scalling down
   * **oc delete nodes worker-N.${cluster_name}.${base_domain}** 

## Create a Glusterfs cluster

To create a Glusterfs cluster, simply specify name, flavor and key. Optionally you can associate it with an existing OpenShift cluster and input a number larger than 3 for works. If an OpenShift cluster is associated, the vms will be deployed into its same subnet so the openshift workers can access it directly.

To create a dynamic storage class, run
```
cat >nfs-pv.yaml <<EOF
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-pv
spec:
  capacity:
    storage: 100Gi
  accessModes:
  - ReadWriteMany
  nfs:
    path: /opt/$cluster_name/data
    server: 192.168.91.8
  persistentVolumeReclaimPolicy: Recycle
EOF
oc create -f nfs-pv.yaml
```

To make the storage class as default storage class, run
```
kubectl patch storageclass glusterfs -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

**Note**: The Gluster service requires a newer version fuse than OpenShift CoreOS has, so when the volumes are created and pvc bounded, they are likely not able to mount and you have to logon one of the Gluster worker nodes and run
```
for vol in `gluster volume list`; do
    gluster volume set $vol ctime off
done
```
