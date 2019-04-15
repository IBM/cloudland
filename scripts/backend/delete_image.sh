#!/bin/bash
cd `dirname $0`
source ../cloudrc
img_name=$1

if [ -f "$cache_dir/${img_name}" ]; then
    rm -f $cache_dir/${img_name}
fi

