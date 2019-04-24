#ifndef _PRAGMA_COPYRIGHT_
#define _PRAGMA_COPYRIGHT_
#pragma comment(copyright, "%Z% %I% %W% %D% %T%\0")
#endif /* _PRAGMA_COPYRIGHT_ */
/****************************************************************************

* Copyright (c) 2008, 2010 IBM Corporation.
* All rights reserved. This program and the accompanying materials
* are made available under the terms of the Eclipse Public License v1.0s
* which accompanies this distribution, and is available at
* http://www.eclipse.org/legal/epl-v10.html

 Classes: SshFunc

 Description: ssh functions
   
 Author: Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 tuhongj      Initial code (D16661)

****************************************************************************/
#ifndef _SSHFUNC_H_
#define _SSHFUNC_H_

#include <stdlib.h>

#include "psec.h"

class SshFunc
{
    public:
        typedef int (psec_set_auth_module_hndlr)(char *, char *, char *, unsigned int *);  
        typedef int (psec_get_id_token_hndlr)(unsigned int, char *, char *, psec_idbuf_t); 
        typedef int (psec_verify_id_token_hndlr)(unsigned int, char *, psec_idbuf_t);    
        typedef int (psec_get_id_from_token_hndlr)(unsigned int, psec_idbuf_t, char *, size_t *);           
        typedef int (psec_free_id_token_hndlr)(unsigned int, psec_idbuf_t); 
        typedef int (psec_get_key_from_token_hndlr)(unsigned int, char *, psec_idbuf_t,	char *,	size_t *);
        typedef int (psec_sign_data_hndlr)(unsigned int, char *, size_t, struct iovec *, int, struct iovec *);
        typedef int (psec_verify_data_hndlr)(unsigned int, char *, size_t, struct iovec *, int, struct iovec *);
        typedef int (psec_free_signature_hndlr)(unsigned int, struct iovec *);

    private:
        void *dlopen_file;
        unsigned int mdlhndl;
        char session_key[64];
        size_t key_len;
        struct iovec user_token;
        
        psec_set_auth_module_hndlr *set_auth_module_hndlr;
        psec_get_id_token_hndlr *get_id_token_hndlr;
        psec_verify_id_token_hndlr *verify_id_token_hndlr;
        psec_get_id_from_token_hndlr *get_id_from_token_hndlr;
        psec_free_id_token_hndlr *free_id_token_hndlr;
        psec_get_key_from_token_hndlr *get_key_from_token_hndlr;
        psec_sign_data_hndlr *sign_data_hndlr;
        psec_verify_data_hndlr *verify_data_hndlr;
        psec_free_signature_hndlr * free_signature_hndlr;

    private:
        SshFunc();
        static SshFunc *instance;
        int set_auth_module(char *name, char *fpath, char *opts);
        int get_sizes(char *fmt);
        bool sshAuth;
        
    public:
        ~SshFunc();
        static SshFunc *getInstance();
        int load(char *libPath = NULL);

        char *get_session_key() { return session_key; }
        size_t get_key_len() {return key_len; }
        int get_id_token(char *tname, char *thost, psec_idbuf_t idtok);
        int verify_id_token(char *uname, psec_idbuf_t idtok);
        int get_id_from_token(psec_idbuf_t idtok, char *usrid, size_t *usridlen);
        int free_id_token(psec_idbuf_t id);
        int get_key_from_token(char *uname, psec_idbuf_t idtok , char *key, size_t *keylen);
        int sign_data(char *key, size_t keylen, struct iovec *inbufs, int num_bufs, struct iovec *sigbufs);
        int verify_data(char *key, size_t keylen, struct iovec *inbufs, int num_bufs, struct iovec *sigbufs);
        int free_signature(struct iovec *sigbufs);

        int sign_data(char *key, size_t keylen, char *bufs[], int sizes[], int num_bufs, struct iovec *sigbufs);
        int verify_data(char *key, size_t keylen, char *bufs[], int sizes[], int num_bufs, struct iovec *sigbufs);

        struct iovec & get_token() { return user_token; }
        int set_session_key(struct iovec *sskey);
        int set_user_token(struct iovec *token);
        int sign_data(struct iovec *inbufs, int num_bufs, struct iovec *sigbufs);
        int verify_data(struct iovec *inbufs, int num_bufs, struct iovec *sigbufs);
        int sign_data(char *bufs[], int sizes[], int num_bufs, struct iovec *sigbufs);
        int verify_data(char *bufs[], int sizes[], int num_bufs, struct iovec *sigbufs);
        int sign_data(struct iovec *sigbufs, int num_bufs, ...);
        int verify_data(struct iovec *sigbufs, int num_bufs, ...);
        int sign_data(char *key, size_t keylen, struct iovec *sigbufs, int num_bufs, ...);
        int verify_data(char *key, size_t keylen, struct iovec *sigbufs, int num_bufs, ...);
        int sign_data(char *key, size_t klen, struct iovec *sigbufs, char *fmt, ...);
        int verify_data(char *key, size_t klen, struct iovec *sigbufs, char *fmt, ...);
};

#define SSHFUNC SshFunc::getInstance()
#define psec_sign_data(sigbufs, ...)        SSHFUNC->sign_data(SSHFUNC->get_session_key(), SSHFUNC->get_key_len(), sigbufs, __VA_ARGS__)
#define psec_verify_data(sigbufs, ...)      SSHFUNC->verify_data(SSHFUNC->get_session_key(), SSHFUNC->get_key_len(), sigbufs, __VA_ARGS__)
#define psec_free_signature(sign)           SSHFUNC->free_signature(sign)

#endif

