#!/bin/bash

export JAEGER_AGENT_PORT=6381
export JAEGER_AGENT_HOST=10.171.202.199
sidecar service install
