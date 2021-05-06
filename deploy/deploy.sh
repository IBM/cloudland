#!/bin/bash

if [[ $# -ne 0 && $# -ne 4 ]]; then 
    echo "Usage: $0 or $0 <admin_password> <database_password> <url_of_conf.json> <if_reboot_kvm_compute>"
    echo "Example: $0, or $0 123456 123456 'file:///tmp/conf.json' [yes|no]"
    exit -1
fi

auto=0
admin_passwd="passw0rd"
db_passwd="passw0rd"
reboot_kvm_compute=0

cland_root_dir=/opt/cloudland
conf=$cland_root_dir/deploy/conf.json

if [[ $# -eq 4 ]]; then
    curl -s -o "$conf" "$3"
    if [[ $? -ne 0 ]]; then
        echo "Check the url of the conf.json"
        exit -1
    fi

    auto=1
    admin_passwd="$1"
    db_passwd="$2"
    reboot_kvm_compute=1
fi

user=`whoami`
if [ $user != "cland" ]; then
    echo "Use user 'cland' to deploy CloudLand."
    exit -1
fi

# check keys
cland_ssh_dir=/home/cland/.ssh
if [ ! -f $cland_ssh_dir/cland.key -o ! -f $cland_ssh_dir/cland.key.pub ]; then
    echo "Please check /home/cland/cland.key and /home/cland/cland.key.pub. Make sure they have been added to all the nodes' authorized_keys"
    exit -1
fi

# check repo
packages="ansible wget jq net-tools gnutls-utils iptables iptables-services postgresql postgresql-server postgresql-contrib"
echo "Checking following prerequisite packages and install them via yum if necessary: "
echo "$packages"
for i in $packages
do
    echo "Checking $i ..."
    rpm -q $i > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        sudo yum install -y $i > /dev/null 2>&1
        if [ $? -ne 0 ]; then
            echo "Install $i failed. Please check yum and re-run the deployment."
            exit -1
        fi
    fi
done

# link backend (default to kvm)
cd $cland_root_dir/scripts
rm -f backend 
ln -s kvm backend

cd $cland_root_dir

echo "Deploying CloudLand ..."

# check configuration file
if [ ! -e $conf ]; then
    echo "No configuration file $cland_root_dir/deploy/conf.json" 
    echo "Create the configuration file according to $cland_root_dir/deploy/conf.json.sample. "
    echo "Re-run the deployment when the configuration file is ready."
    exit -1
fi

# prepare hosts/hosts
mkdir -p $cland_root_dir/deploy/hosts
hosts=$cland_root_dir/deploy/hosts/hosts

# process controller
controller=$(jq -r .controller < $conf)
hname=$(echo $controller | jq -r .hostname)
ip=$(echo $controller | jq -r .ip)
cat > $hosts <<EOF
[imgrepo]
$hname ansible_host=$ip ansible_ssh_private_key_file=$cland_ssh_dir/cland.key

[cland]
$hname ansible_host=$ip ansible_ssh_private_key_file=$cland_ssh_dir/cland.key

[web]
$hname ansible_host=$ip ansible_ssh_private_key_file=$cland_ssh_dir/cland.key

[database]
$hname ansible_host=$ip ansible_ssh_private_key_file=$cland_ssh_dir/cland.key

[hyper]
EOF

new_conf="yes"
if [ ! -e "/opt/cloudland/web/clui/conf/config.toml" ]; then
    if [[ $auto -eq 0 ]]; then
        read -s -p "Set the 'admin' login password: " admin_passwd
        echo
        read -s -p "Set the database login password: " db_passwd
        echo
    fi
else 
    new_conf="no"
    admin_passwd=$(grep 'password = ' /opt/cloudland/web/clui/conf/config.toml | awk '{print $3}' | tr -d '"')
    db_passwd=$(grep 'user=postgres' /opt/cloudland/web/clui/conf/config.toml | awk '{print $6}' | awk -F '=' '{print $2}')
fi

cd $cland_root_dir/deploy

# deploy controller (base database, web and cland)
ansible-playbook controller.yml -e "admin_passwd=$admin_passwd db_passwd=$db_passwd new_conf=$new_conf" --tags hosts,selinux,imgrepo,database,web,console,fe_srv,firewall

# process compute nodes
compute=$(jq -r .compute < $conf)
length=$(echo $compute | jq length)
let end=length-1
if [ $end -lt 0 ]; then
    ansible-playbook service.yml --tags start_cloudland
else
    ./deploy_compute.sh 0 $end $reboot_kvm_compute
fi
echo "Done."
