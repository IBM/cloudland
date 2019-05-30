#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 2 ] && die "$0 <ID> <format>"

ID=$1
format=$2

image=$image_cache/image-${ID}.${format}
rm -f $image
