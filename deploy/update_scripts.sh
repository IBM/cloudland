#!/bin/bash

user=`whoami`
if [ $user != "cland" ]; then
    echo "Use user 'cland' to add compute node."
    exit -1
fi

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

cd $cland_root_dir/deploy

# controller
controller=$(jq -r .controller < $conf)
hname=$(echo $controller | jq -r .hostname)

# compute nodes
compute=$(jq -r .compute < $conf)
length=$(echo $compute | jq length)
let end=length-1
if [ $end -ge 0 ]; then
    for ((i=0;i<=$end;i++));
    do
        node=$(echo $compute | jq '.['$i']')
        ansible-playbook compute.yml -e "$node" -e "controller=$hname" --tags scripts_only
    done
fi