#!/bin/bash

memory=$(cat /proc/meminfo | grep MemTotal | awk '{print $2}')
let hugepage=$memory/2048*3/4
echo "$hugepage"
