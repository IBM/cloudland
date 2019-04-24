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
#include <stdlib.h>
#include <sys/uio.h>

#ifndef _H_PSEC_H
#define _H_PSEC_H

#ifdef __cplusplus
extern "C" {
#endif

#define PSEC_MEMORY_ERR 1
#define PSEC_MEMORY2_ERR 2
#define PSEC_MEMORY3_ERR 3
#define PSEC_MEMORY4_ERR 4
#define PSEC_MEMORY5_ERR 5
#define PSEC_ARGS_ERR 11
#define PSEC_ARGS2_ERR 12
#define PSEC_ARGS3_ERR 13
#define PSEC_ARGS4_ERR 14
#define PSEC_INTERNAL_ERR 20
#define PSEC_MUTEX_INIT_ERR 30
#define PSEC_MODULE_PTR_ERR 101
#define PSEC_MODULE_PATH_ERR 102
#define PSEC_MODULE_FILE_ERR 103
#define PSEC_MODULE_IREG_ERR 104
#define PSEC_MODULE_SIZE_ERR 105
#define PSEC_MODULE_INIT_ERR 106
#define PSEC_MODULE_OPTS_ERR 107
#define PSEC_MODULE_OPTS2_ERR 108
#define PSEC_MODULE_INTERNAL_ERR 109
#define PSEC_MODULE_HNDL_ERR 111
#define PSEC_MODULE_HNDL2_ERR 112
#define PSEC_DLOPEN_ERR 120
#define PSEC_DLSYM_ERR 121
#define PSEC_DLSYM2_ERR 122
#define PSEC_DLSYM3_ERR 123
#define PSEC_DLSYM4_ERR 124
#define PSEC_DLSYM5_ERR 125
#define PSEC_DLSYM6_ERR 126
#define PSEC_NOT_SUPPORTED_ERR 150

typedef struct iovec psec_idbuf_desc, *psec_idbuf_t;

int
psec_set_auth_module(		/* sets the authentication module params */
	char *,					/* mnemonic--can be NULL */
	char *,					/* file full path name */
	char *,					/* flags--can be NULL */
	unsigned int *);		/* return handle to the authentication method */

int
psec_get_id_token(			/* get a client's identity token */
	unsigned int,			/* handle to the authentication method */
	char *,					/* target user identity */
	char *,					/* target host name */
	psec_idbuf_t);			/* address of identity buffer descriptor */

int
psec_verify_id_token(		/* verifies a client's identity token */
	unsigned int,			/* handle to the authentication method */
	char *,					/* name of target user */
	psec_idbuf_t);			/* address of identity buffer descriptor */

int
psec_get_id_from_token(		/* returns a client's identity */
	unsigned int,			/* handle to the authentication method */
	psec_idbuf_t,			/* address of id token descriptor */
	char *,					/* memory for id string */
	size_t *);				/* provided/required length of id string */

int
psec_free_id_token(			/* frees memory allocated for id token */
	unsigned int,			/* handle to the authentication method */
	psec_idbuf_t);			/* address of id token descriptor */

int
psec_get_key_from_token(	/* returns session key from id token */
	unsigned int,			/* handle to the authentication method */
	char *,					/* name of target user */
	psec_idbuf_t,			/* address of identity buffer descriptor */
	char *,					/* memory for key data */
	size_t *);				/* provided/required length of key data */

int
psec_sign_data(				/* sign data */
	unsigned int,			/* handle to the authentication method */
	char *,					/* key data */
	size_t,					/* length of key data */
	struct iovec *,			/* address of input data vector */
	int,					/* number of vectors */
	struct iovec *);		/* address of signature buffer */

int
psec_verify_data(			/* verify signature */
	unsigned int,			/* handle to the authentication method */
	char *,					/* key data */
	size_t,					/* length of key data */
	struct iovec *,			/* address of input data vector */
	int,					/* number of vectors */
	struct iovec *);		/* address of signature buffer */

int
psec_free_signature(		/* frees memory allocated for signature */
	unsigned int,			/* handle to the authentication method */
	struct iovec *);		/* address of signature buffer */

#ifdef __cplusplus
}
#endif

#endif	// _H_PSEC_H
