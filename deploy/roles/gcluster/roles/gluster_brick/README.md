gluster_brick
=================

This role helps the user to manage bricks of a gluster volume.

Requirements
------------
Ansible version 2.5 or above


Role Variables
--------------
The following are the variables available for this role

| Name | Choices | Default value | Comments |
| --- | --- | --- | --- |
| gluster_cluster_new_bricks | | UNDEF | Contains the  list of bricks along with the new bricks to be added to the GlusterFS volume. The format of the bricks is mountpoint/brick_dir |
| gluster_cluster_hosts | | UNDEF | Contains the list of hosts that have to be peer probed. |
| gluster_cluster_volume | | glustervol | Name of the volume. Refer GlusterFS documentation for valid characters in a volume name. |
| gluster_cluster_remove_bricks | | UNDEF | Contains the list of bricks to be removed. |

### Tags
--------
cluster_brick

### Example Playbook
--------------------

Add a brick to an existing gluster volume

```yaml
---
- name: Add bricks to an existing volume
  hosts: gluster_servers
  remote_user: root
  gather_facts: false

  vars:
    gluster_cluster_hosts:
      - 10.70.42.83
      - 10.70.42.85
    gluster_cluster_volume: testvol
    gluster_cluster_new_bricks: '/mnt/brick1/b1,/mnt/brick2/b2,/mnt/brick1/b2'

  roles:
    - gluster_brick

```

The above playbook will be run as part of gluster.cluster. However if you
want to run just the gluster_brick role use the tag cluster_brick.

For example:
\# ansible-playbook -i inventory_file playbook_file.yml --tags cluster_brick

License
-------

GPLv3

