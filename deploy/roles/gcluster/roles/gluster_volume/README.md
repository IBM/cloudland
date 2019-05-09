gluster_volume
==============

This role helps the user to manage a gluster volume.
Volume can be of any type - replicate, distributed-replicate, arbiter ...

Requirements
------------
Ansible version 2.5 or above


Role Variables
--------------

The following are the variables available for this role

| Name | Choices | Default value | Comments |
| --- | --- | --- | --- |
| gluster_cluster_arbiter_count | | UNDEF | Number of arbiter bricks to use (Only for arbiter volume types). |
| gluster_cluster_bricks | | UNDEF | Brick paths on servers. Multiple brick paths can be separated by commas. |
| gluster_cluster_disperse_count | | UNDEF | Disperse count for the volume. If this value is specified, a dispersed volume will be  created |
| gluster_cluster_force | **yes** / **no** | no | Force option will be used while creating a volume, any warnings will be suppressed. |
| gluster_cluster_hosts | | | Contains the list of hosts that have to be peer probed. |
| gluster_cluster_redundancy_count | | UNDEF | Specifies the number of redundant bricks while creating a disperse volume. If redundancy count is missing an optimal value is computed. |
| gluster_cluster_replica_count | **2** / **3** | UNDEF | Replica count while creating a volume. Currently replica 2 and replica 3 are supported. |
| gluster_cluster_state | **present** / **absent** / **started** / **stopped** / **set** | present | If value is present volume will be created. If value is absent, volume will be deleted. If value is started, volume will be started. If value is stopped, volume will be stopped. |
| gluster_cluster_transport | **tcp** / **rdma** / **tcp,rdma** | tcp | The transport type for the volume. |
| gluster_cluster_volume | | glustervol | Name of the volume. Refer GlusterFS documentation for valid characters in a volume name. |


### Variables specific to the respective volume type
-----------------------------------------------------

1. #### Arbitrated-Replicated Volume
| Name | Choices | Default value | Comments |
| --- | --- | --- | --- |
| gluster_cluster_replica_count | **2** / **3** | UNDEF | Replica count while creating a volume. Currently replica 2 and replica 3 are supported. |
| gluster_cluster_arbiter_count | | UNDEF | Number of arbiter bricks to use (Only for arbiter volume types). |

2. #### Distributed-Replicated Volume
| Name | Choices | Default value | Comments |
| --- | --- | --- | --- |
| gluster_cluster_replica_count | **2** / **3** | UNDEF | Replica count while creating a volume. Currently replica 2 and replica 3 are supported. |

3. #### Distributed-Dispersed Volume
| Name | Choices | Default value | Comments |
| --- | --- | --- | --- |
| gluster_cluster_disperse_count | | UNDEF | Disperse count for the volume. If this value is specified, a dispersed volume will be  created |
| gluster_cluster_redundancy_count | | UNDEF | Specifies the number of redundant bricks while creating a disperse volume. If redundancy count is missing an optimal value is computed. |

### Tags
--------
cluster_volume

### Example Playbook
--------------------

Create a GlusterFS volume


```yaml
---
- name: Create Gluster cluster
  hosts: gluster_servers
  remote_user: root
  gather_facts: false

  vars:
    gluster_cluster_hosts:
      - 10.70.41.212
      - 10.70.42.156
    gluster_cluster_volume: testvol
    gluster_cluster_force: 'yes'
    gluster_cluster_bricks: '/mnt/brick1/b1,/mnt/brick1/b2'

  roles:
    - gluster.cluster

```

The above playbook will be run as part of gluster.cluster. However if you
want to run just the volume_create role use the tag volume_create.

For example:
\# ansible-playbook -i inventory_file playbook_file.yml --tags cluster_volume

License
-------

GPLv3

