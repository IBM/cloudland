#!/bin/bash

flock -w 300 /tmp/db.lock sqlite3 /opt/cloudland/db/benchmark.db "delete from benchmark"
date
for ID in {1..10000}; do
    flock -w 300 /tmp/db.lock sqlite3 /opt/cloudland/db/benchmark.db "insert into benchmark (id) values ('$ID')"  
#    echo $ID: $(date) >> /tmp/benchmark.txt
    /opt/cloudland/bin/sendmsg "inter" "/opt/cloudland/scripts/backend/`basename $0` $ID"
done
date
