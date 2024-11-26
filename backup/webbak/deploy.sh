#!/bin/bash
export CLADMIN_PWD=${CLADMIN_PWD:-/var/local/cladmin}
export CLADMIN_PID=${CLADMIN_PID:-$(pidof cladmin)}
export RELEASE_NAME=${RELEASE_NAME:-cladmin}
export RELEASE_VERSION=${RELEASE_VERSION:-latest}
localip=$(ip -4 addr show dev eth0|grep inet | cut -d/ -f1 | awk '{print $2}')
export CLADMIN_ADMIN_LISTEN=${CLADMIN_ADMIN_LISTEN:-${localip}:50080}
startsh=$(cd $(dirname ${BASH_SOURCE[0]}); pwd)/start.sh
nohup $startsh &>/dev/null &
