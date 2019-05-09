firewall_config
===============

This role helps the user to configure the firewall.


Requirements
------------
Ansible version 2.5 or above

Role Variables
--------------

The following are the variables available for this role

### firewall_config
| Name                     |Choices| Default value         | Comments                          |
|--------------------------|-------|-----------------------|-----------------------------------|
| gluster_infra_fw_state | enabled / disabled / present / absent    | UNDEF   | Enable or disable a setting. For ports: Should this port accept(enabled) or reject(disabled) connections. The states "present" and "absent" can only be used in zone level operations (i.e. when no other parameters but zone and state are set). |
| gluster_infra_fw_ports |    | UNDEF    | A list of ports in the format PORT/PROTO. For example 111/tcp. This is a list value.  |
| gluster_infra_fw_permanent  | true/false  | true | Whether to make the rule permanenet. |
| gluster_infra_fw_zone    | work / drop / internal/ external / trusted / home
/ dmz/ public / block | public   | The firewalld zone to add/remove to/from |
| gluster_infra_fw_services |    | UNDEF | Name of a service to add/remove to/from firewalld - service must be listed in output of firewall-cmd --get-services. This is a list variable|

### Tags
--------
firewall

### Example Playbook
--------------------

Configure the ports and services related to GlusterFS:


```yaml
---
- name: Setting up backend
  remote_user: root
  hosts: gluster_servers
  gather_facts: false

  vars:
     # Firewall setup
     gluster_infra_fw_ports:
        - 2049/tcp
        - 54321/tcp
        - 5900/tcp
        - 5900-6923/tcp
        - 5666/tcp
        - 16514/tcp
     gluster_infra_fw_permanent: true
     gluster_infra_fw_state: enabled
     gluster_infra_fw_zone: public
     gluster_infra_fw_services:
        - glusterfs

  roles:
     - gluster.infra

```

The above playbook will be run as part of the gluster.infra. However if you
want to run just the firewall role use the tag firewall.

For example:
\# ansible-playbook -i inventory_file playbook_file.yml --tags firewall

License
-------

GPLv3

