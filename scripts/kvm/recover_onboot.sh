#!/bin/bash

cd $(dirname $0)
source ../cloudrc

for inst in $(virsh list --all | grep 'shut off' | awk '{print $2}'); do
    virsh start $inst
done

touch $run_dir/need_to_sync_fdb

exit 0
