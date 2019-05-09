backend_reset
=========

This role unmounts the filesystem and deletes specified logical volumes and volume groups.

Note: The mounted filesystems should not be busy. For example, if the GlusterFS volumes are running, then umount fails and logical volumes will not be deleted.

Requirements
------------

Ansible version 2.5 or above.
VDO utilities (Optional)

Role Variables
--------------

| Name                     |Required?| Default value         | Comments                          |
|--------------------------|-------|-----------------------|-----------------------------------|
| gluster_infra_reset_mnt_paths | No | | Mount point of the brick which has to be unmounted and logical volumes deleted. |
| gluster_infra_reset_volume_groups | No | | Name of the volume group which has to be deleted. All corresponding logical volumes and physical volumes will be deleted. |
| gluster_infra_reset_vdos | No | | Name of the vdo devices that have to be removed |

Example Playbook
----------------

Unmount the filesystem and delete the roles

    - hosts: gluster_servers
      vars:
        - gluster_infra_reset_mnt_path: /mnt/foo
        - gluster_infra_reset_volume_group: gluster_vg
      roles:
         - gluster.infra

License
-------

GPLv3

Author Information
------------------

Sachidananda Urs <surs@redhat.com>
