#!/bin/bash

gcc -g -Wall -Wunused -Werror -o getipaddr getip.c dhcp.c lookup.c
gcc -g -Wall -Wunused -Werror -o askip askip.c
