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

#ifndef _H_PSM_H
#define _H_PSM_H

#ifdef __cplusplus
extern "C" {
#endif

typedef struct iovec psm_idbuf_desc, *psm_idbuf_t;

#define PSM__SUCCESS 0
#define PSM__MEMORY_ERR 1

int
psm__init(				/* one time initialization of the module */
	char *);			/* options */

int
psm__get_id_token(		/* gets a client's identity token */
	char *,				/* target user identity */
	char *,				/* tartet host name */
	psm_idbuf_t);		/* address of identity buffer descriptor */

int
psm__verify_id_token(	/* verifies a client's identity token */
	char *,				/* name of user to authorize as */
	psm_idbuf_t);		/* address of identity buffer descriptor */

int
psm__get_id_from_token(	/* returns a client's identity */
	psm_idbuf_t,		/* address of identity buffer descriptor */
	char *,				/* memory for id string */
	size_t *);			/* provided/required length of the id string */

int
psm__free_id_token(		/* frees memory allocated for id token */
	psm_idbuf_t);		/* address of identity buffer descriptor */

int
psm__get_key_from_token(/* gets session key from token */
	char *,				/* user name */
	psm_idbuf_t,		/* address of identity buffer descriptor */
	unsigned char *,	/* memory for key data */
	size_t *);			/* provided/required length of key data */

int
psm__sign_data(			/* sign data */
	unsigned char *,	/* key data */
	size_t,				/* length of key data */
	struct iovec *,		/* address of input data vector */
	int,				/* number of buffers in vector */
	struct iovec *);	/* address of signature buffer */

int
psm__verify_data(		/* verify signature */
	unsigned char *,	/* key data */
	size_t,				/* length of key data */
	struct iovec *,		/* address of input data vector */
	int,				/* number of buffers in vector */
	struct iovec *);	/* address of signature buffer */

int
psm__free_signature(	/* deallocates mem. allocated for signature */
	struct iovec *);	/* address of signature buffer */

void
psm__cleanup();			/* one time cleanup of the module */

#ifdef __cplusplus
}
#endif

#endif	// _H_PSM_H
