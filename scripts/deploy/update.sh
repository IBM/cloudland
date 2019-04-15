#!/bin/bash

frontend=cloudland
frontha=cloudland-ha
hypers=compute-1:bmhyper

ansible $frontend -s -a 'systemctl stop cloudland'
ansible $frontend -m script -a './wait_term.sh'

ansible $hypers -m copy -a 'src=/opt/cloudland/bin dest=/opt/cloudland'
ansible $hypers -m copy -a 'src=/opt/cloudland/lib64 dest=/opt/cloudland'

ansible $frontend -s -a 'systemctl start cloudland'
sleep 5
ansible $frontha -s -a 'systemctl stop cloudland'
ansible $frontha -m script -a './wait_term.sh'
ansible $frontha -m copy -a 'src=/opt/cloudland/bin dest=/opt/cloudland'
ansible $frontha -m copy -a 'src=/opt/cloudland/lib64 dest=/opt/cloudland'
ansible $hypers -m shell -a 'cp -rf /opt/cloudland/bin/* /opt/cloudland/bin-1/'
ansible $hypers -m shell -a 'cp -rf /opt/cloudland/lib64/* /opt/cloudland/lib64-1/'
