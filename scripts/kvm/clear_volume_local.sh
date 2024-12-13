#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 2 ] && die "$0 <vol_ID> <vol_UUID> <path>"

vol_ID=$1
vol_UUID=$2
path=$3

rm -f $volume_dir/$path
