#!/bin/bash

bs_inventory=/opt/deploy/bs_inventory

function bootstrap() {
    local ip=$1
    local hostname=$2
    local passwd=$(slcli hardware detail $hostname --passwords | grep "^users.*root" | awk '{print $3}')
    ssh-keygen -R $ip >/dev/null 2>&1
    cat > $bs_inventory/$hostname << EOF
[bootstrap]
$hostname ansible_host=$ip ansible_user=root ansible_ssh_pass=$passwd
EOF
}

mkdir $bs_inventory
slcli hardware list | grep val | while read line; do
    name=$(echo $line | awk '{print $2}')
    ip=$(echo $line | awk '{print $4}')
    bootstrap $ip $name
done
