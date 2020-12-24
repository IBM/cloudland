#!/bin/bash

cd $(dirname $0)
source ../cloudrc

vm_ID=$(printf $guest_userid_template $1)
action=$2
state=error

# only support "start" and "stop" now
if [ $action = "start" ] || [ $action = "shutdown" ]; then
    [ $action = "shutdown" ] && action=softstop
    rc=$(curl -s $zvm_service/guests/$vm_ID/action -X POST -d '{"action":"'"$action"'"}' | jq .rc)
    if [ $rc -eq 0 ]; then
        if [ $action = "start" ]; then
            state=running
        else
            state=shut_off
        fi
    fi
fi

echo "|:-COMMAND-:| $(basename $0) '$1' '$state'"
