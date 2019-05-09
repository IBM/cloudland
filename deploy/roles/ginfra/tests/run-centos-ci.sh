#!/bin/bash

# setting up virtual environment
yum -y install epel-release
yum -y install libselinux-python yum-utils \
  device-mapper-persistent-data \
  lvm2 gcc python-virtualenv
virtualenv --system-site-packages env
source env/bin/activate

# install dependency packages
pip install ansible molecule docker-py

# prerequisites to install docker
yum-config-manager \
    --add-repo \
    https://download.docker.com/linux/centos/docker-ce.repo
yum -y install docker-ce

# start and enable Docker service
systemctl start docker
systemctl enable docker

# run tests
cd gluster-ansible-infra/roles/firewall_config/  
molecule create
molecule test

cd gluster-ansible-infra/roles/backend_setup/
molecule create
molecule test
