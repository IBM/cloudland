#!/bin/bash

cd `dirname $0`
source ../cloudrc

ID=$1

flock -w 300 /tmp/db.lock sqlite3 /opt/cloudland/db/benchmark.db "update benchmark set end=datetime('now') where id=$ID"
#echo $ID: $(date) >> /tmp/benchmark.txt
