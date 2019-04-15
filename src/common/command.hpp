/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#ifndef _COMMAND_HPP
#define _COMMAND_HPP

struct __attribute__((__packed__)) Command {
    int id;
    int extra;
    char *control;
    char *content;
};

#endif
