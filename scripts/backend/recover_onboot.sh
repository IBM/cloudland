#!/bin/bash
#!/bin/bash

cd $(dirname $0)
source ../cloudrc

for conf in $cache_dir/router/*; do
    ./load_keepalived_conf.py -q $conf/keepalived.conf
done

for inst in $(virsh list --all | grep 'shut off' | awk '{print $2}'); do
    virsh start $inst
done
