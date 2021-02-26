#!/bin/bash

passwd=$1

su - postgres -c "psql -c \"alter user postgres with password '$passwd'\""
su - postgres -c "psql -U postgres -tc \"SELECT 1 FROM pg_database WHERE datname = 'hypercube'\"" | grep -q 1 || su - postgres -c "psql -c 'create database hypercube'"
