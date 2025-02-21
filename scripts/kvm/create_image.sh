#!/bin/bash

cd $(dirname $0)
source ../cloudrc

async_exec ./async_job/$(basename $0) $*
