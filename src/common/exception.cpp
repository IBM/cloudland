/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#include <assert.h>

#include "exception.hpp"

const char * ErrMsg[] = {
    "SCI initialize failed.",
    "SCI broadcast failed.",
    "SCI unicast failed.",
    "Socket bind failed.",
    "Socket recv failed.",
    "Socket send failed.",
    "Bad address.",
    "Invalid user.",
    "System call failed."
};

CommonException::CommonException(int code) throw()
        : errCode(code)
{
}

const char * CommonException::getErrMsg() const throw()
{
    return ErrMsg[errCode];
}

int CommonException::getErrCode() const throw()
{
    return errCode;
}

