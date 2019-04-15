/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#ifndef _COMMON_EXCEPTION_HPP
#define _COMMON_EXCEPTION_HPP

class CommonException {
    public:
        enum CODE {
            SCI_INIT_ERROR,
            SCI_BCAST_ERROR,
            SCI_UCAST_ERROR,
            SOCK_BIND_ERROR,
            SOCK_RECV_ERROR,
            SOCK_SEND_ERROR,
            BAD_ADDR_INFO,
            INVALID_USER,
            SYS_CALL
        };
        
    private:
        int        errCode;
        
    public:
        CommonException(int code) throw();
        
        const char * getErrMsg() const throw();
        int getErrCode() const throw();
};

#endif

