gluster.infra
=========

This role helps the user to setup the backend for GlusterFS filesystem.

backend_setup:
        - Create volume groups, logical volumes (thinpool, thin lv, thick lv)
        - Create xfs filesystem
        - Mount the filesystem

Requirements
------------

Ansible version 2.5 or above


Role Variables
--------------

### backend_setup
-----------------
| Name                     |Choices| Default value         | Comments                          |
|--------------------------|-------|-----------------------|-----------------------------------|
| gluster_infra_vdo || UNDEF | Mandatory argument if vdo has to be setup. Key/Value pairs have to be given. name and device are the keys, see examples for syntax. |
| gluster_infra_disktype | JBOD / RAID6 / RAID10  | UNDEF   | Backend disk type. |
| gluster_infra_diskcount || UNDEF | RAID diskcount, can be ignored if disktype is JBOD  |
| gluster_infra_volume_groups  || | Mandatory variable, key/value pairs of vgname and pvname. pvname can be comma-separated values if more than a single pv is needed for a vg. See below for syntax. |
| gluster_infra_stripe_unit_size || UNDEF| Stripe unit size (KiB). *DO NOT* including trailing 'k' or 'K'  |
| gluster_infra_thinpools || | Thinpool data. This is a dictionary with keys vgname, thinpoolname, thinpoolsize, and poolmetadatasize. See below for syntax and example. |
| gluster_infra_lv_logicalvols || UNDEF | This is a list of hash/dictionary variables, with keys, lvname and lvsize. See below for example. |
| gluster_infra_thick_lvs || UNDEF | Optional. Needed only if thick volume has to be created. This is a dictionary with vgname, lvname, and size as keys. See below for example |
| gluster_infra_mount_devices | | UNDEF | This is a dictionary with mount values. path, vgname, and lv are the keys. |
| gluster_infra_cache_vars | | UNDEF | This variable contains list of dictionaries for setting up LV cache. Variable has following keys: vgname, cachedisk, cachethinpoolname, cachelvname, cachelvsize, cachemetalvname, cachemetalvsize, cachemode. The keys are explained in more detail below|


#### VDO Variable
------------
If the backend disk has to be configured with VDO the variable gluster_infra_vdo has to be defined.

| Name                     |Choices| Default value         | Comments                          |
|--------------------------|-------|-----------------------|-----------------------------------|
| gluster_infra_vdo || UNDEF | Mandatory argument if vdo has to be setup. Key/Value pairs have to be given. See below for syntax and list of keys and values supported. |


```
For Example:
gluster_infra_vdo:
   - { name: 'hc_vdo_1', device: '/dev/vdb' }
   - { name: 'hc_vdo_2', device: '/dev/vdc' }
   - { name: 'hc_vdo_3', device: '/dev/vdd' }
```

The gluster_infra_vdo variable supports the following keys:

| VDO Key                 | Default value         | Comments                          |
|--------------------------|-----------------------|-----------------------------------|
| state | present | VDO state, if present VDO will be created, if absent VDO will be deleted. |
| activated | 'yes' | Whether VDO has to be activated upon creation. |
| running | 'yes' | Whether VDO has to be started |
| logicalsize | UNDEF | Logical size for the vdo |
| compression | 'enabled' | Whether compression has to be enabled |
| blockmapcachesize | '128M' | The amount of memory allocated for caching block map pages, in megabytes (or may be issued with an LVM-style suffix of K, M, G, or T). The default (and minimum) value is 128M. |
| readcache | 'disabled' | Enables or disables the read cache. |
| readcachesize | 0 | Specifies the extra VDO device read cache size in megabytes. |
| emulate512 | 'off' | Enables 512-byte emulation mode, allowing drivers or filesystems to access the VDO volume at 512-byte granularity, instead of the default 4096-byte granularity. |
| slabsize | '2G' | The size of the increment by which the physical size of a VDO volume is grown, in megabytes (or may be issued with an LVM-style suffix of K, M, G, or T). Must be a power of two between 128M and 32G. |
| writepolicy | 'sync' | Specifies the write policy of the VDO volume. The 'sync' mode acknowledges writes only after data is on stable storage. |
| indexmem | '0.25' | Specifies the amount of index memory in gigabytes. |
| indexmode | 'dense' | Specifies the index mode of the Albireo index. |
| ackthreads | '1' | Specifies the number of threads to use for acknowledging completion of requested VDO I/O operations. Valid values are integer values from 1 to 100 (lower numbers are preferable due to overhead). The default is 1.|
| biothreads | '4' | Specifies the number of threads to use for submitting I/O operations to the storage device. Valid values are integer values from 1 to 100 (lower numbers are preferable due to overhead). The default is 4. |
| cputhreads | '2' | Specifies the number of threads to use for CPU-intensive work such as hashing or compression. Valid values are integer values from 1 to 100 (lower numbers are preferable due to overhead). The default is 2. |
| logicalthreads | '1' | Specifies the number of threads across which to subdivide parts of the VDO processing based on logical block addresses. Valid values are integer values from 1 to 100 (lower numbers are preferable due to overhead). The default is 1.|
| physicalthreads | '1' | Specifies the number of threads across which to subdivide parts of the VDO processing based on physical block addresses. Valid values are integer values from 1 to 16 (lower numbers are preferable due to overhead). The physical space used by the VDO volume must be larger than (slabsize * physicalthreads). The default is 1. |


#### Volume Groups variable
------------------------
| Name                     |Choices| Default value         | Comments                          |
|--------------------------|-------|-----------------------|-----------------------------------|
| gluster_infra_volume_groups || UNDEF | This is a list of hash/dictionary variables, with keys, vgname and pvname. See below for example. |

```
For Example:
gluster_infra_volume_groups:
   - { vgname: 'volgroup1', pvname: '/dev/sdb' }
   - { vgname: 'volgroup2', pvname: '/dev/mapper/vdo_device1' }
   - { vgname: 'volgroup3', pvname: '/dev/sdc,/dev/sdd'
```

#### Logical Volume variable
-----------------------
| Name                     |Choices| Default value         | Comments                          |
|--------------------------|-------|-----------------------|-----------------------------------|
| gluster_infra_lv_logicalvols || UNDEF | This is a list of hash/dictionary variables, with keys,  lvname, vgname, thinpool, and lvsize. See below for example. |

```
For Example:
gluster_infra_lv_logicalvols:
   - { vgname: 'vg_vdb', thinpool: 'foo_thinpool', lvname: 'vg_vdb_thinlv', lvsize: '500G' }
   - { vgname: 'vg_vdc', thinpool: 'bar_thinpool', lvname: 'vg_vdc_thinlv', lvsize: '500G' }
```

#### Thick LV variable
-----------------------
| Name                     |Choices| Default value         | Comments                          |
|--------------------------|-------|-----------------------|-----------------------------------|
| gluster_infra_thick_lvs || UNDEF | This is a list of hash/dictionary variables, with keys: vgname, lvname, and size. See below for example. |


```
For Example:
gluster_infra_thick_lvs:
   - { vgname: 'vg_ssd', lvname: 'thick_lv_1', size: '500G' }
   - { vgname: 'vg_sdc', lvname: 'thick_lv_2', size: '100G' }
```

#### Thinpool variable
----------------------
| Name                     |Choices| Default value         | Comments                          |
|--------------------------|-------|-----------------------|-----------------------------------|
| gluster_infra_thinpools || UNDEF | This is a list of hash/dictionary variables, with keys: vgname, thinpoolname, thinpoolsize, and poolmetadatasize. See below for example. |

```
gluster_infra_thinpools:
  - {vgname: 'vg_vdb', thinpoolname: 'foo_thinpool', thinpoolsize: '10G', poolmetadatasize: '1G' }
  - {vgname: 'vg_vdc', thinpoolname: 'bar_thinpool', thinpoolsize: '20G', poolmetadatasize: '1G' }
```

* poolmetadatasize: Metadata size for LV, recommended value 16G is used by default. That value can be overridden by setting the variable. Include the unit [G\|M\|K]
* thinpoolname: Can be ignored, a name is formed using the given vgname followed by '_thinpool'
* vgname: Name of the volume group this thinpool should belong to.

#### Variables for setting up cache
-----------------------------------------
| Name                     |Choices| Default value         | Comments                          |
|--------------------------|-------|-----------------------|-----------------------------------|
| gluster_infra_cache_vars | | UNDEF | This is a dictionary with keys: vgname, cachedisk, cachethinpoolname, cachelvname, cachelvsize, cachemetalvname, cachemetalvsize, cachemode |

```
vgname - The vg which will be extended to setup cache.
cachedisk - The SSD disk which will be used to setup cache. Complete path, for eg: /dev/sdd
cachethinpoolname - The existing thinpool on the volume group mentioned above.
cachelvname - Logical volume name for setting up cache, an lv with this name is created.
cachelvsize - Cache logical volume size
cachemetalvname - Meta LV name.
cachemetalvsize - Meta LV size
cachemode - Cachemode, default is writethrough.
```

For example:
```
   - { vgname: 'vg_vdb', cachedisk: '/dev/vdd', cachethinpoolname: 'foo_thinpool', cachelvname: 'cachelv', cachelvsize: '20G', cachemetalvname: 'cachemeta', cachemetalvsize: '100M', cachemode: 'writethrough' }
```


#### Variables for mounting the filesystem
-----------------------------------------
| Name                     |Choices| Default value         | Comments                          |
|--------------------------|-------|-----------------------|-----------------------------------|
| gluster_infra_mount_devices | | UNDEF | This is a dictionary with mount values. path, vgname, and lv are the keys. |

```
For example:
gluster_infra_mount_devices:
        - { path: '/mnt/thinv', vgname: <vgname>, lv: <lvname> }
        - { path: '/mnt/thicklv', vgname: <vgname>, lv: 'thick_lv_1' }
```


Example Playbook
----------------

Configure the ports and services related to GlusterFS, create logical volumes and mount them.


```
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
     gluster_infra_fw_services:
        - glusterfs

     # Set a disk type, Options: JBOD, RAID6, RAID10
     gluster_infra_disktype: RAID6

     # RAID6 and RAID10 diskcount (Needed only when disktype is raid)
     gluster_infra_diskcount: 10
     # Stripe unit size always in KiB
     gluster_infra_stripe_unit_size: 128

     # Variables for creating volume group
     gluster_infra_volume_groups:
       - { vgname: 'vg_vdb', pvname: '/dev/vdb' }
       - { vgname: 'vg_vdc', pvname: '/dev/vdc' }

     # Create a thick volume name
     gluster_infra_thick_lvs:
       - { vgname: 'vg_vdb', lvname: 'thicklv_1', size: '100G' }

     # Create thinpools
     gluster_infra_thinpools:
       - {vgname: 'vg_vdb', thinpoolname: 'foo_thinpool', thinpoolsize: '100G', poolmetadatasize: '16G'}
       - {vgname: 'vg_vdc', thinpoolname: 'bar_thinpool', thinpoolsize: '500G', poolmetadatasize: '16G'}

     # Create a thin volume
     gluster_infra_lv_logicalvols:
       - { vgname: 'vg_vdb', thinpool: 'foo_thinpool', lvname: 'vg_vdb_thinlv', lvsize: '500G' }
       - { vgname: 'vg_vdc', thinpool: 'bar_thinpool', lvname: 'vg_vdc_thinlv', lvsize: '500G' }

     # Mount the devices
     gluster_infra_mount_devices:
       - { path: '/mnt/thicklv', vgname: 'vg_vdb', lvname: 'thicklv_1' }
       - { path: '/mnt/thinlv1', vgname: 'vg_vdb', lvname: 'vg_vdb_thinlv' }
       - { path: '/mnt/thinlv2', vgname: 'vg_vdc', lvname: 'vg_vdc_thinlv' }

  roles:
     - gluster.infra
```

See also: https://github.com/gluster/gluster-ansible-infra/tree/master/playbooks


License
-------

GPLv3
