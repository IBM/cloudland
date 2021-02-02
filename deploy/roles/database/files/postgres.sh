#!/bin/bash

su - postgres -c "psql -U postgres -tc \"SELECT 1 FROM pg_database WHERE datname = 'hypercube'\"" | grep -q 1 || su - postgres -c "psql -c 'create database hypercube'"
