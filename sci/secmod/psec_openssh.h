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

 Description: psec functions from poe
   
 Author: Serban Maerean

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 serban      Initial code (D153875)

****************************************************************************/

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif
#include "psec_mod.h"

#ifndef _H_PSEC_OSSH_H
#define _H_PSEC_OSSH_H

#define OSSH_MEMORY_ERR 1
#define OSSH_PWUID_ERR 2
#define OSSH_PWUID_HOMEDIR_ERR 3
#define OSSH_IDTOK_INVALID_ERR 4
#define OSSH_IDTOK_NOTVERIFD_ERR 5
#define OSSH_IDTOK_SKEW_ERR 6
#define OSSH_IDTOK_USER_ERR 7
#define OSSH_UNAME_INVALID_ERR 8
#define OSSH_DSA_INVALID_FORMAT_ERR 9
#define OSSH_RSA_INVALID_FORMAT_ERR 10
#define OSSH_CIPHER_VERIF_ERR 11
#define OSSH_NOUSERID_ERR 12
#define OSSH_USERID_LEN_ERR 13
#define OSSH_ARG_ERR 14
#define OSSH_IDTOK_NOSKEY_ERR 15

#define OSSH_CFGFILE_ERR 20
#define OSSH_CFGFILE_SIZE_ERR 21
#define OSSH_CFGFILE_DATA_ERR 22
#define OSSH_CFGFILE_OPEN_ERR 23
#define OSSH_CFGFILE_GETS_ERR 24
#define OSSH_IDFILE_NAME_ERR 25
#define OSSH_IDFILE_PATH_ERR 26
#define OSSH_IDFILE_SIZE_ERR 27
#define OSSH_IDFILE_OPEN_ERR 28
#define OSSH_IDFILE_READ_ERR 29
#define OSSH_IDFILE_DEFS_ERR 30
#define OSSH_PIDFILE_PATH_ERR 31
#define OSSH_PIDFILE_SIZE_ERR 32
#define OSSH_PIDFILE_OPEN_ERR 33
#define OSSH_PIDFILE_GETS_ERR 34
#define OSSH_AUTHZFILE_NAME_ERR 35
#define OSSH_AUTHZFILE_PATH_ERR 36
#define OSSH_AUTHZFILE_OPEN_ERR 37
#define OSSH_AUTHZFILE_GETS_ERR 38

#define OSSH_DSA_NEW_ERR 40
#define OSSH_RSA_NEW_ERR 41
#define OSSH_DRSA_SIZE_ERR 42
#define OSSH_DRSA_SIGN_ERR 43
#define OSSH_DSA_VERIFY_ERR 44
#define OSSH_RSA_VERIFY_ERR 45
#define OSSH_SHA1_ERR 46
#define OSSH_BIO_NEW_MEMBUF_ERR 47
#define OSSH_BIO_NEW_ERR 48
#define OSSH_BIO_PUSH_ERR 49
#define OSSH_BIO_READ_ERR 50
#define OSSH_BN_BIN2BN_ERR 51 
#define OSSH_BN_DEC2BN_ERR 52
#define OSSH_RSA_PUBLIC_ENCRYPT_ERR 53
#define OSSH_RSA_PRIVATE_DECRYPT_ERR 54
#define OSSH_SIG_INVALID_ERR 55
#define OSSH_DATA_INVALID_ERR 56
#define OSSH_MD5_INIT_ERR 57
#define OSSH_MD5_UPDATE_ERR 57
#define OSSH_MD5_FINAL_ERR 58

#define OSSH_PTHRD_SETSPECIFIC_ERR 60
#define OSSH_PTHRD_KEYCREATE_ERR 61
#define OSSH_DLSYM_ERR 62
#define OSSH_DLOPEN_ERR 63

#define OSSH_SKEY_NOTRSA_ERR 70
#define OSSH_SKEY_NOKEY_ERR 71
#define OSSH_AES_KEYSCHED_ERR 72
#define OSSH_AES_ENCRYPT_ERR 73
#define OSSH_SIG_SIGNATURE_ERR 74

#endif	// _H_PSEC_OSSH_H
