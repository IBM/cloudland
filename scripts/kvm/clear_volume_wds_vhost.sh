#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 2 ] && die "$0 <vol_ID> <path>"

vol_ID=$1
path=$2

rm -f $volume_dir/$path
