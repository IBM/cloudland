#!/bin/bash
mkdir -p /opt/cloudland{bin,cache,db,etc,lib64,log,run,scripts,xml}
mkdir -p /opt/cloudland/cache/{xml,meta,instance}
chown -R ubuntu.ubuntu /opt/cloudland
