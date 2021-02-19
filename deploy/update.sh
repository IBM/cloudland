#!/bin/bash

cland_root_dir=/opt/cloudland
cd $(dirname $0)

function update_web()
{
    cd $cland_root_dir/web/clui
    go build
    sudo systemctl restart hypercube
}

function update_cland()
{
    cd $cland_root_dir/src
    export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
    make clean
    make
    sudo systemctl stop cloudland
    make install
    sudo systemctl start cloudland
}

function update_sci()
{
    cd $cland_root_dir/sci
    make
    sudo systemctl stop scid
    sudo make install 
    sudo systemctl start scid
}

function sync_hyper()
{
    cd $cland_root_dir/deploy
    ansible hyper -b -a 'systemctl stop cloudlet'
    ansible hyper -b -a 'systemctl stop scid'
    ansible-playbook cloudland.yml --tags sync,be_pkg,firewall
    ansible hyper -b -a 'systemctl start scid'
    ansible hyper -b -a 'systemctl start cloudlet'
}

git pull
update_web
update_sci
update_cland
sync_hyper
