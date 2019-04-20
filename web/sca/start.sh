#!/bin/bash
cd ${CLADMIN_PWD}
sleep 1
kill -9 $(pidof cladmin)
tar Cxfz / releases/${RELEASE_NAME}/${RELEASE_VERSION}/cladmin.tgz
exec /usr/local/bin/cladmin serve
