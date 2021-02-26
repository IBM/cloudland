#!/bin/bash

user=`whoami`
if [ $user != "cland" ]; then
    echo "Use user 'cland' to add compute node."
    exit -1
fi

if [ $# -ne 2 -o $2 -lt $1 ]; then
    echo "deploy_compute.sh <begin> <end>"
    exit -1
fi

begin=$1
end=$2

# set cland root dir
cland_root_dir=/opt/cloudland

# check configuration file
conf=$cland_root_dir/deploy/conf.json
if [ ! -e $conf ]; then
    echo "No configuration file $cland_root_dir/deploy/conf.json" 
    echo "Create the configuration file according to $cland_root_dir/deploy/conf.json.sample. "
    echo "Re-run the deployment when the configuration file is ready."
    exit -1
fi

compute=$(jq -r .compute < $conf)
length=$(echo $compute | jq length)
if [ $end -ge $length ]; then
    echo "Only $length nodes available. Check the range of the nodes to deploy."
    exit -1
fi

cland_ssh_dir=/home/cland/.ssh
hosts=$cland_root_dir/deploy/hosts/hosts

# add nodes to hosts
for ((i=$begin;i<=$end;i++));
do
    node=$(echo $compute | jq '.['$i']')
    id=$(echo $node | jq -r .id)
    if [ $id != $i ]; then
        echo "ID set error: $id vs. $i"
        echo "Compute ID ranges [0, 1, 2, ...] Check the configuraiton and re-run the deployment."
        exit -1
    else
        hname=$(echo $node | jq -r .hostname)
        ip=$(echo $node | jq -r .ip)
        virt_type=$(echo $node | jq -r .virt_type)
        zone_name=$(echo $node | jq -r .zone_name)

        # check if the node already exists, append it to hosts if it does not exist
        entry="$hname ansible_host=$ip ansible_ssh_private_key_file=$cland_ssh_dir/cland.key client_id=$id zone_name=$zone_name virt_type=$virt_type"
        if [ $(grep -c "client_id=$id" $hosts) -eq 0 ]; then
            cat >> $hosts <<EOF
$entry
EOF
        else
            if [ $(grep -c "$entry" $hosts) -ne 0 ]; then
                echo "Node $hname ClientID=$id already exists. Update this compute node."
            else
                echo "ClientID=$id exists but it seems that its entry doesn't match the configuration. Check the conf.json and hosts/hosts and re-run the depolyment."
                exit -1
            fi
        fi
    fi
done

cd $cland_root_dir/deploy

# generate /etc/hosts, ./etc/host.list and restart cloudland
ansible-playbook service.yml --tags hosts,selinux,gen_host_list,start_cloudland

sleep 1s

# restart compute nodes' cloudlet after cloudland
ansible-playbook service.yml --tags start_cloudlet

# deploy compute
controller=$(jq -r .controller < $conf)
hname=$(echo $controller | jq -r .hostname)
for ((i=$begin;i<=$end;i++));
do
    node=$(echo $compute | jq '.['$i']')
    ansible-playbook compute.yml -e "$node" -e "controller=$hname" --tags hyper
done