#!/bin/bash
qemu-img create -f qcow2 -b \
/opt/cloudland/cache/image/image-1.qcow2 \
/opt/cloudland/cache/image/image-2.qcow2

virt-install \
--name=image-2 \
--vcpus=1 \
--memory=1024 \
--disk path=/opt/cloudland/cache/image/image-2.qcow2,bus=virtio,format=qcow2 \
--network bridge=virbr0 \
