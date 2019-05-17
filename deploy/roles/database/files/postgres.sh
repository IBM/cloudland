#!/bin/bash

passwd=$1

su - postgres -c "psql -c \"alter user postgres with password '$passwd'\""
su - postgres -c "psql -c 'create database hypercube'"
