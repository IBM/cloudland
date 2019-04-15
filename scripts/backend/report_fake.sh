#!/bin/bash

cd `dirname $0`
source ../cloudrc

let cpu=$RANDOM%32+32
let memory=$RANDOM%128000+128000
let disk=$RANDOM%1000+1000
network=0
load=0
echo "cpu=$cpu/$cpu memory=$memory/$memory disk=$disk/$disk network=$network/$network load=$load/$load"
