#!/bin/bash

cd `dirname $0`
source ../cloudrc
[ $# -lt 1 ] && die "$0 <mac_addr>"
mac_addr=$1
sudo /usr/sbin/bridge fdb del $mac_addr dev $zlayer2_interface