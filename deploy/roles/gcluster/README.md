gluster.cluster
===============

This role helps the user to set up a GlusterFS cluster, manage gluster volumes and peer operations.
The gluster.cluster role has multiple sub-roles which are invoked depending on the variables that are set.
The sub-roles are:

  1. gluster_volume - manage gluster volumes(create, delete, start, stop, set)
  2. gluster_brick - perform add/remove/replace operations on existing bricks

Requirements
------------
- Ansible version 2.5 or above.
- GlusterFS version 3.2 or above.


Role Variables
--------------

These are the superset of role_varibles. They are explained further in the respective sub-roles directory.

### gluster_volume
------------------

| Name | Choices | Default value | Comments |
| --- | --- | --- | --- |
| gluster_cluster_arbiter_count | | UNDEF | Number of arbiter bricks to use (Only for arbiter volume types). |
| gluster_cluster_bricks | | UNDEF | Bricks that form the GlusterFS volume. The format of the bricks would be hostname:mountpoint/brick_dir alternatively user can provide just mountpoint/birck_dir, in such a case gluster_hosts variable has to be set |
| gluster_cluster_disperse_count | | UNDEF | Disperse count for the volume. If this value is specified, a dispersed volume will be  created |
| gluster_cluster_force | **yes** / **no** | no | Force option will be used while creating a volume, any warnings will be suppressed. |
| gluster_cluster_hosts | | | Contains the list of hosts that have to be peer probed. |
| gluster_cluster_redundancy_count | | UNDEF | Specifies the number of redundant bricks while creating a disperse volume. If redundancy count is missing an optimal value is computed. |
| gluster_cluster_replica_count | **2** / **3** | UNDEF | Replica count while creating a volume. Currently replica 2 and replica 3 are supported. |
| gluster_cluster_state | **present** / **absent** / **started** / **stopped** / **set** | present | If value is present volume will be created. If value is absent, volume will be deleted. If value is started, volume will be started. If value is stopped, volume will be stopped. |
| gluster_cluster_transport | **tcp** / **rdma** / **tcp,rdma** | tcp | The transport type for the volume. |
| gluster_cluster_volume | | glustervol | Name of the volume. Refer GlusterFS documentation for valid characters in a volume name. |

### gluster_brick
-----------------

| Name | Choices | Default value | Comments |
| --- | --- | --- | --- |
| gluster_cluster_new_bricks | | UNDEF | Contains the  list of bricks along with the new bricks to be added to the GlusterFS volume. The format of the bricks is mountpoint/brick_dir |
| gluster_cluster_hosts | | UNDEF | Contains the list of hosts that have to be peer probed. |
| gluster_cluster_volume | | glustervol | Name of the volume. Refer GlusterFS documentation for valid characters in a volume name. |
| gluster_cluster_remove_bricks | | UNDEF | Contains the list of bricks to be removed. |


### Example Playbook
--------------------

Create a GlusterFS volume and set specific options.

```yaml
---
- name: Create Gluster cluster
  hosts: gluster_servers
  remote_user: root
  gather_facts: false

  vars:
    # gluster volume
    gluster_cluster_hosts:
      - 10.70.41.212
      - 10.70.42.156
    gluster_cluster_volume: testvol
    gluster_cluster_transport: 'tcp'
    gluster_cluster_force: 'yes'
    gluster_cluster_bricks: '/mnt/brick1/b1,/mnt/brick1/b2'

    # variables to create specific type of the volume
    gluster_cluster_replica_count: 3
    gluster_cluster_arbiter_count: 1

    # variables to set specific volume options
    gluster_cluster_options: 'performance.cache-size:256MB'

  roles:
    - gluster.cluster

```

License
-------

GPLv3
