#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 1 ] && die "$0 <vol_ID> <vol_UUID> <path>"

wds_vol_ID=$2

get_wds_token
wds_curl DELETE "api/v2/sync/block/volumes/$wds_vol_ID?force=false"
