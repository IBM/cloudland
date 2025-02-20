#!/bin/bash

cd $(dirname $0)
source ../cloudrc

async_exec ./create_image_async.sh $*
