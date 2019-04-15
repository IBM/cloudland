#!/bin/bash
cd "$(dirname $0)"
source ../cloudrc
echo "|:-COMMAND-:| `basename $0` '$SCI_CLIENT_ID' '$SCI_PARENT_HOSTNAME' '$SCI_PARENT_ID'"
