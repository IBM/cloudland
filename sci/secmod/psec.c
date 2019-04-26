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
#include <pthread.h>
#include <stdlib.h>
#include <stdio.h>
#include <errno.h>
#include <string.h>
#include <dlfcn.h>
#include <fcntl.h>
#include <sys/stat.h>

#include "psec.h"
#include "psec_lib.h"
#include "psec_mod.h"

char *_psec_rtne_list[PSM_SYMTBLE_SIZE] = {
#define PSM__INIT 0
	"psm__init",
#define PSM__CLEANUP 1
	"psm__cleanup",
#define PSM__GET_ID_TOKEN 2
	"psm__get_id_token",
#define PSM__VERIFY_ID_TOKEN 3
	"psm__verify_id_token",
#define PSM__GET_ID_FROM_TOKEN 4
	"psm__get_id_from_token",
#define PSM__FREE_ID_TOKEN 5
	"psm__free_id_token",
#define PSM__GET_KEY_FROM_TOKEN 6
	"psm__get_key_from_token",
#define PSM__SIGN_DATA 7
	"psm__sign_data",
#define PSM__VERIFY_DATA 8
	"psm__verify_data",
#define PSM__FREE_SIGNATURE 9
	"psm__free_signature"};


_psec_state_desc _PSEC_STATE = {0, PTHREAD_MUTEX_INITIALIZER, 0, NULL, NULL};

int
_psec_load_auth_module(
	_psec_module_t psmp)
{
	int rc = 0;
do {
	if (!psmp) {
		// printf("Error [%s:d]: internal error: invalid auth module_pinter\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_PTR_ERR; break;
	}
	pthread_mutex_lock(&psmp->psm_mutex);
	pthread_cleanup_push((void(*)(void *))pthread_mutex_unlock, (void *)&psmp->psm_mutex);
do {
	if (PSM_STATE_INITED&psmp->psm_stindex) break;  // already init'ed
do {
	if (PSM_STATE_LOADED&psmp->psm_stindex) break;	// already loaded
	if (!psmp->psm_fpath) {
		// printf("Error [%s:%d]: internal error: invalid module file path (%s)\n", __FILE__, __LINE__, psmp->psm_fpath);
		rc = PSEC_MODULE_PATH_ERR; break;
	}
	if (!(psmp->psm_object = dlopen(psmp->psm_fpath, RTLD_NOW))) {
		// char *errmsg = dlerror();
		// printf("Error [%s:%d]: dlopen() failed: %s\n", __FILE__, __LINE__, errmsg?errmsg:"<no error message>");
		rc = PSEC_DLOPEN_ERR; break;
	}
do {
	// resolve all the symbols
	int i; for (i=0; i<PSM_SYMTBLE_SIZE; i++) {
		if (!(psmp->psm_symtble[i] = dlsym(psmp->psm_object, _psec_rtne_list[i]))) {
			// char *errmsg = dlerror();
			// printf("Error [%s:%d]: dlsym() failed: %s\n", __FILE__, __LINE__, errmsg?errmsg:"<no error message>");
			if (i<6) rc = PSEC_DLSYM_ERR+i;
			break;
		}
	}
	if (rc) {
		memset(psmp->psm_symtble, 0, sizeof(psmp->psm_symtble));
		break;
	}
} while (0);
	if (rc) {
		dlclose(psmp->psm_object); psmp->psm_object = NULL;
		break;
	}
	psmp->psm_stindex |= PSM_STATE_LOADED;
} while (0);
	if (rc) break;
	pthread_cleanup_push((void (*)(void *))(psmp->psm_symtble[PSM__CLEANUP]), NULL);
do {
	// call the module's initialization routine
	if (rc = ((int(*)())psmp->psm_symtble[PSM__INIT])(psmp->psm_opts)) {
		((void (*)())psmp->psm_symtble[PSM__CLEANUP])();
		// printf("Error [%s:%d]: init failed w/rc = %d for module %s\n", __FILE__, __LINE__, rc, psmp->psm_fpath);
		rc = PSEC_MODULE_INIT_ERR; break;
	}
	psmp->psm_stindex |= PSM_STATE_INITED;
} while (0);
	pthread_cleanup_pop(0);
}  while (0);
	pthread_cleanup_pop(1);
} while (0);
	return rc;
}

int
psec_set_auth_module(
	char *name,
	char *fpath,
	char *opts,
	unsigned int *mdlhndl)
{
	int rc = 0, tmdlhndl;
do {
	// check arguments
	if (!fpath) { rc = PSEC_ARGS_ERR; break; }
	if ('/'!=fpath[0]) { rc = PSEC_MODULE_PATH_ERR; break; }
{
	struct stat amsbuf = {0};
	if (0>stat(fpath, &amsbuf)) {
		// printf("Error [%s:%d]: stat() failed w/errno = %d for %s\n", __FILE__, __LINE__, errno, fpath);
		rc = PSEC_MODULE_FILE_ERR; break;
	}
	if (!S_ISREG(amsbuf.st_mode)) { rc = PSEC_MODULE_IREG_ERR; break; }
	if (0==amsbuf.st_size) { rc = PSEC_MODULE_SIZE_ERR; break; }
}
{	// get the mutex lock and load the authentication module 
	_psec_module_t psmp = NULL;
	pthread_mutex_lock(&_PSEC_STATE.pss_mutex);
	pthread_cleanup_push((void(*)(void *))pthread_mutex_unlock, &_PSEC_STATE.pss_mutex);
do {
	// search for an existing entry
	psmp = _PSEC_STATE.pss_modules;
	while (psmp) {
		if (!strcmp(psmp->psm_fpath, fpath)) break;
		psmp = psmp->next;
	}
	if (psmp) break;		// module data already allocated
	// allocate memory for the module's internal data structure
	psmp = malloc(sizeof(_psec_module_sec));
	if (!psmp) {
		// printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		rc = PSEC_MEMORY_ERR; break;
	}
do {
	memset(psmp, 0, sizeof(*psmp));
	if (rc = pthread_mutex_init(&psmp->psm_mutex, NULL)) {
		// printf("Error [%s:%d]: pthread_mutex_init() failed w/rc = %d\n", __FILE__, __LINE__, rc);
		rc = PSEC_MUTEX_INIT_ERR; break;
	}
do {
	if (!(psmp->psm_fpath = strdup(fpath))) {
		// printf("Error [%s:%d]: strdup() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		rc = PSEC_MEMORY2_ERR; break;
	}
do {
	if (name&&!(psmp->psm_name = strdup(name))) {
		// printf("Error [%s:%d] strdup() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		rc = PSEC_MEMORY3_ERR; break;
	}
do {
do {	// translate options
	// options are a list of comma separated characters (no spaces allowed);
	// options can have values within square brackets, immediately
	// following the character identifying the option.
	// one important optiom is 'm' and it defines the options that the
	// shared library will pass to the auth module's init function
	// if there is no value to this option, the auth module's init function
	// will get a NULL pointer
	char *cp = opts;
	if (!cp||('\0'==*cp)) break;
	do {
		switch (*cp) {
			case 'm': case 'M':
				if (psmp->psm_opts) {
					// printf("Error [%s:%d]: module options can be specified only once\n", __FILE__, __LINE__);
					rc = PSEC_MODULE_OPTS_ERR; break;
				}
				if (('\0'==*++cp)||('['!=*cp)) break;		// no value
{				// find the end of module options string
				char *endcp = cp++;
				while (*++endcp) {
					int noofbkslshs = 0;
					if (!(endcp = strchr(endcp, ']'))) {
						// printf("Error [%s:%d]: no closing bracket in module options\n", __FILE__, __LINE__);
						rc = PSEC_MODULE_OPTS2_ERR; break;
					}
{					// check whether quote is back-slashed
					char *tcp = endcp;
					while ('\\' == *--tcp) noofbkslshs++;
}
					if (0==noofbkslshs%2) break; // non-back-slashed quote
				}
				if (0<endcp-cp) {
					if (!(psmp->psm_opts = calloc(endcp-cp+1, sizeof(char)))) {
						// printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
						rc = PSEC_MEMORY4_ERR; break;
					}
					strncpy(psmp->psm_opts, cp, endcp-cp);
				}
				cp = endcp;
}
				break;
			default:
				// ignore unknown options
				break;
		}
		if (rc) break;
	} while (*++cp);
} while (0);
	if (rc) break;
{	// add the module's data structure to the list
	_psec_module_t *tmdlslist = NULL;
	if (!(tmdlslist = realloc(_PSEC_STATE.pss_mdlslist, (_PSEC_STATE.pss_modcnt+1)*sizeof(_psec_module_t)))) {
		// printf("Error [%s:%d]: realloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		rc = PSEC_MEMORY5_ERR; break;
	}
	_PSEC_STATE.pss_mdlslist = tmdlslist;
}
	_PSEC_STATE.pss_mdlslist[_PSEC_STATE.pss_modcnt++] = psmp;
	psmp->psm_stindex = (0xffff&_PSEC_STATE.pss_modcnt)<<16;
	__add_elem_to_dllist((__dlink_elem_t)psmp, (__dlink_elem_t *)&_PSEC_STATE.pss_modules);
} while (0);
	if (rc&&psmp->psm_opts) free(psmp->psm_opts);
	if (rc&&psmp->psm_name) free(psmp->psm_name);
} while (0);
	if (rc) free(psmp->psm_fpath);
} while (0);
	if (rc) pthread_mutex_destroy(&psmp->psm_mutex);
} while (0);
	if (rc) free(psmp);
} while (0);
	pthread_cleanup_pop(1);
	if (rc) break;
	tmdlhndl = (0xffff0000&psmp->psm_stindex)>>16;
	if (!(PSM_STATE_INITED&psmp->psm_stindex)) {
		if (rc = _psec_load_auth_module(psmp)) {
			// printf("Error [%s:%d]: failed to load and init auth module (%s)\n", __FILE__, __LINE__, psmp->psm_fpath);
			break;
		}
	}
}
	*mdlhndl = tmdlhndl;
} while (0);
	return rc;
}

int
psec_get_id_token(
	unsigned int mdlhndl,
	char *tname,
	char *thost,
	psec_idbuf_t idtok)
{
	int rc = 0;
do {
	// check arguments
	if (!idtok) {
		// printf("Error [%s:%d]: invalid id token argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS_ERR; break;
	}
	// find the authentication module by handle
	if ((1>mdlhndl)||(_PSEC_STATE.pss_modcnt<mdlhndl)) {
		// printf("Error [%s:%d]: invalid module handle\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL_ERR; break;
	}
{
	_psec_module_t psmp = _PSEC_STATE.pss_mdlslist[mdlhndl-1];
	if (!psmp||(mdlhndl!=((0xfff000&psmp->psm_stindex)>>16))) {
		// this should not happen, something went very wrong
		// printf("Error [%s:%d]: internal failure: no auth module\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL2_ERR; break;
	}
	// check the module's state
	if (!(PSM_STATE_INITED&psmp->psm_stindex)) {
		if (rc = _psec_load_auth_module(psmp)) {
			// printf("Error [%s:%d]: failed to load and init auth module (%s)\n", __FILE__, __LINE__, psmp->psm_fpath);
			break;
		}
	}
	// call the module's routine
	if (rc = ((int(*)(char*, char *, psec_idbuf_t))psmp->psm_symtble[PSM__GET_ID_TOKEN])(tname, thost, idtok)) {
		// printf("Error [%s:%d]: auth module's get_id_token() failed w/rc = %d\n", __FILE__, __LINE__, rc);
		rc = PSEC_MODULE_INTERNAL_ERR; break;
	}
}
} while (0);
	return rc;
}

int
psec_verify_id_token(
	unsigned int mdlhndl,
	char *uname,
	psec_idbuf_t idtok)
{
	int rc = 0;
do {
	// check arguments
	if (!idtok) {
		// printf("Error [%s:%d]: invalid id token argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS_ERR; break;
	}
	// find the authentication module by handle
	if ((1>mdlhndl)||(_PSEC_STATE.pss_modcnt<mdlhndl)) {
		// printf("Error [%s:%d]: invalid module handle\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL_ERR; break;
	}
{
	_psec_module_t psmp = _PSEC_STATE.pss_mdlslist[mdlhndl-1];
	if (!psmp||(mdlhndl!=((0xfff000&psmp->psm_stindex)>>16))) {
		// this should not happen, something went very wrong 
		// printf("Error [%s:%d]: internal failure: no auth module\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL2_ERR; break;
	}
	// check the module's state
	if (!(PSM_STATE_INITED&psmp->psm_stindex)) {
		if (rc = _psec_load_auth_module(psmp)) {
			// printf("Error [%s:%d]: failed to load and init auth module (%s)\n", __FILE__, __LINE__, psmp->psm_fpath);
			break;
		}
	}
	// call the module's routine
	if (rc = ((int(*)(char *, psec_idbuf_t))psmp->psm_symtble[PSM__VERIFY_ID_TOKEN])(uname, idtok)) {
		// printf("Error [%s:%d]: auth module's verify_id_token() failed w/rc = %d\n", __FILE__, __LINE__, rc);
		rc = PSEC_MODULE_INTERNAL_ERR; break;
	}
}
} while (0);
	return rc;
}

int
psec_get_id_from_token(
	unsigned int mdlhndl,
	psec_idbuf_t idtok,
	char *usrid,
	size_t *usridlen)
{
	int rc = 0;
do {
	// check arguments
	if (!idtok) {
		// printf("Error [%s:%d]: invalid id token argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS_ERR; break;
	}
	if (!usridlen) {
		// printf("Error [%s:%d]: invalid id length argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS3_ERR; break;
	}
	if ((0!=*usridlen)&&(!usrid)) {
		// printf("Error [%s:%d]: invalid id argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS2_ERR; break;
	}
	// find the authentication module by handle
	if ((1>mdlhndl)||(_PSEC_STATE.pss_modcnt<mdlhndl)) {
		// printf("Error [%s:%d]: invalid module handle\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL_ERR; break;
	}
{
	_psec_module_t psmp = _PSEC_STATE.pss_mdlslist[mdlhndl-1];
	if (!psmp||(mdlhndl!=((0xfff000&psmp->psm_stindex)>>16))) {
		// this should not happen, something went very wrong
		// printf("Error [%s:%d]: internal failure: no auth module\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL2_ERR; break;
	}
	// check the module's state
	if (!(PSM_STATE_INITED&psmp->psm_stindex)) {
		if (rc = _psec_load_auth_module(psmp)) {
			// printf("Error [%s:%d]: failed to load and init auth module (%s)\n", __FILE__, __LINE__, psmp->psm_fpath);
			break;
		}
	}
	// call the module's routine
	if (rc = ((int(*)(psec_idbuf_t, char*, size_t *))psmp->psm_symtble[PSM__GET_ID_FROM_TOKEN])(idtok, usrid, usridlen)) {
		// printf("Error [%s:%d]: auth module's get_id_from_token() failed w/rc = %d\n", __FILE__, __LINE__, rc);
		if (PSM__MEMORY_ERR == rc) { rc = PSEC_MEMORY_ERR; break; }
		rc = PSEC_MODULE_INTERNAL_ERR; break;
	}
}
} while (0);
	return rc;
}

int
psec_get_key_from_token(
	unsigned int mdlhndl,
	char *uname,
	psec_idbuf_t idtok,
	char *key,
	size_t *keylen)
{
	int rc = 0;
do {
	// check arguments
	if (!idtok) {
		// printf("Error [%s:%d]: invalid id token argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS_ERR; break;
	}
	if (!keylen) {
		// printf("Error [%s:%d]: invalid key length argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS3_ERR; break;
	}
	if ((0!=*keylen)&&(!key)) {
		// printf("Error [%s:%d]: invalid key argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS2_ERR; break;
	}
	// find the authentication module by handle
	if ((1>mdlhndl)||(_PSEC_STATE.pss_modcnt<mdlhndl)) {
		// printf("Error [%s:%d]: invalid module handle\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL_ERR; break;
	}
{
	_psec_module_t psmp = _PSEC_STATE.pss_mdlslist[mdlhndl-1];
	if (!psmp||(mdlhndl!=((0xfff000&psmp->psm_stindex)>>16))) {
		// this should not happen, something went very wrong
		// printf("Error [%s:%d]: internal failure: no auth module\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL2_ERR; break;
	}
	// check the module's state
	if (!(PSM_STATE_INITED&psmp->psm_stindex)) {
		if (rc = _psec_load_auth_module(psmp)) {
			// printf("Error [%s:%d]: failed to load and init auth module (%s)\n", __FILE__, __LINE__, psmp->psm_fpath);
			break;
		}
	}
	// check whether routine supported by authentication module
	if (!psmp->psm_symtble[PSM__GET_KEY_FROM_TOKEN]) {
		// printf("Error [%s:%d]: function not supported\n", __FILE__, __LINE__);
		rc = PSEC_NOT_SUPPORTED_ERR; break;
	}
	// call the module's routine
	if (rc = ((int(*)(char *, psec_idbuf_t, char *, size_t *))psmp->psm_symtble[PSM__GET_KEY_FROM_TOKEN])(uname, idtok, key, keylen)) {
		// printf("Error [%s:%d]: auth module's get_key_from_token() failed w/rc = %d\n", __FILE__, __LINE__, rc);
		if (PSM__MEMORY_ERR == rc) { rc = PSEC_MEMORY_ERR; break; }
		rc = PSEC_MODULE_INTERNAL_ERR; break;
	}
}
} while (0);
	return rc;
}

int
psec_sign_data(
	unsigned int mdlhndl,
	char *key,
	size_t keylen,
	struct iovec *in,
	int cnt,
	struct iovec *signature)
{
	int rc = 0;
do {
	// check arguments
	if (!key) {
		// printf("Error [%s:%d]: invalid key argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS_ERR; break;
	}
	if (0==keylen) {
		// printf("Error [%s:%d]: invalid key length argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS2_ERR; break;
	}
	if ((NULL==in)||(0==in->iov_len)||(NULL==in->iov_base)) {
		// printf("Error [%s:%d]: invalid input data argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS3_ERR; break;
	}
	if (NULL==signature) {
		// printf("Error [%s:%d]: invalid signature argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS4_ERR; break;
	}
	// find the authentication module by handle
	if ((1>mdlhndl)||(_PSEC_STATE.pss_modcnt<mdlhndl)) {
		// printf("Error [%s:%d]: invalid module handle\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL_ERR; break;
	}
{
	_psec_module_t psmp = _PSEC_STATE.pss_mdlslist[mdlhndl-1];
	if (!psmp||(mdlhndl!=((0xfff000&psmp->psm_stindex)>>16))) {
		// this should not happen, something went very wrong
		// printf("Error [%s:%d]: internal failure: no auth module\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL2_ERR; break;
	}
	// check the module's state
	if (!(PSM_STATE_INITED&psmp->psm_stindex)) {
		if (rc = _psec_load_auth_module(psmp)) {
			// printf("Error [%s:%d]: failed to load and init auth module (%s)\n", __FILE__, __LINE__, psmp->psm_fpath);
			break;
		}
	}
	// check whether routine is supported by authentication module
	if (!psmp->psm_symtble[PSM__SIGN_DATA]) {
		// printf("Error [%s:%d]: function not supported\n", __FILE__, __LINE__);
		rc = PSEC_NOT_SUPPORTED_ERR; break;
	}
	// call the module's routine
	if (rc = ((int(*)(char*, size_t, struct iovec*, int, struct iovec*))psmp->psm_symtble[PSM__SIGN_DATA])(key, keylen, in, cnt, signature)) {
		// printf("Error [%s:%d]: auth module's sign_data() failed w/rc = %d\n", __FILE__, __LINE__, rc);
		if (PSM__MEMORY_ERR == rc) { rc = PSEC_MEMORY_ERR; break; }
		rc = PSEC_MODULE_INTERNAL_ERR; break;
	}
}
} while (0);
	return rc;
}

int
psec_verify_data(
	unsigned int mdlhndl,
	char *key,
	size_t keylen,
	struct iovec *in,
	int cnt,
	struct iovec *signature)
{
	int rc = 0;
do {
	// check arguments
	if (!key) {
		// printf("Error [%s:%d]: invalid key argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS_ERR; break;
	}
	if (0==keylen) {
		// printf("Error [%s:%d]: invalid key length argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS2_ERR; break;
	}
	if ((NULL==in)||(0==cnt)) {
		// printf("Error [%s:%d]: invalid input data argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS3_ERR; break;
	}
	if ((NULL==signature)||(0==signature->iov_len)||(NULL==signature->iov_base)) {
		// printf("Error [%s:%d]: invalid signature argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS4_ERR; break;
	}
	// find the authentication module by handle
	if ((1>mdlhndl)||(_PSEC_STATE.pss_modcnt<mdlhndl)) {
		// printf("Error [%s:%d]: invalid module handle\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL_ERR; break;
	}
{
	_psec_module_t psmp = _PSEC_STATE.pss_mdlslist[mdlhndl-1];
	if (!psmp||(mdlhndl!=((0xfff000&psmp->psm_stindex)>>16))) {
		// this should not happen, something went very wrong
		// printf("Error [%s:%d]: internal failure: no auth module\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL2_ERR; break;
	}
	// check the module's state
	if (!(PSM_STATE_INITED&psmp->psm_stindex)) {
		if (rc = _psec_load_auth_module(psmp)) {
			// printf("Error [%s:%d]: failed to load and init auth module (%s)\n", __FILE__, __LINE__, psmp->psm_fpath);
			break;
		}
	}
	// check whether routine is supported by authentication module
	if (!psmp->psm_symtble[PSM__VERIFY_DATA]) {
		// printf("Error [%s:%d]: function not supported\n", __FILE__, __LINE__);
		rc = PSEC_NOT_SUPPORTED_ERR; break;
	}
	// call the module's routine
	if (rc = ((int(*)(char*, size_t, struct iovec*, int, struct iovec*))psmp->psm_symtble[PSM__VERIFY_DATA])(key, keylen, in, cnt, signature)) {
		// printf("Error [%s:%d]: auth module's verify_data() failed w/rc = %d\n", __FILE__, __LINE__, rc);
		if (PSM__MEMORY_ERR == rc) { rc = PSEC_MEMORY_ERR; break; }
		rc = PSEC_MODULE_INTERNAL_ERR; break;
	}
}
} while (0);
	return rc;

}

int
psec_free_id_token(
	unsigned int mdlhndl,
	psec_idbuf_t idtok)
{
	int rc = 0;
do {
	// check arguments
	if (!idtok) {
		// printf("Error [%s:%d]: invalid id token argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS_ERR; break;
	}
	// find the authentication module by handle
	if ((1>mdlhndl)||(_PSEC_STATE.pss_modcnt<mdlhndl)) {
		// printf("Error [%s:%d]: invalid module handle\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL_ERR; break;
	}
{
	_psec_module_t psmp = _PSEC_STATE.pss_mdlslist[mdlhndl-1];
	if (!psmp||(mdlhndl!=((0xfff000&psmp->psm_stindex)>>16))) {
		// this should not happen, something went very wrong
		// printf("Error [%s:%d]: internal failure: no auth module\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL2_ERR; break;
	}
	// check the module's state
	if (!(PSM_STATE_INITED&psmp->psm_stindex)) {
		if (rc = _psec_load_auth_module(psmp)) {
			// printf("Error [%s:%d]: failed to load and init auth module (%s)\n", __FILE__, __LINE__, psmp->psm_fpath);
			break;
		}
	}
	// call the module's routine
	if (rc = ((int(*)(psec_idbuf_t))psmp->psm_symtble[PSM__FREE_ID_TOKEN])(idtok)) {
		// printf("Error [%s:%d]: auth module's free_id_token() failed w/rc = %d\n", __FILE__, __LINE__, rc);
		rc = PSEC_MODULE_INTERNAL_ERR; break;
	}
}
} while (0);
	return rc;
}

int
psec_free_signature(
	unsigned int mdlhndl,
	struct iovec *signature)
{
	int rc = 0;
do {
	// check arguments
	if (!signature) {
		// printf("Error [%s:%d]: invalid signature argument\n", __FILE__, __LINE__);
		rc = PSEC_ARGS_ERR; break;
	}
	// find the authentication module by handle
	if ((1>mdlhndl)||(_PSEC_STATE.pss_modcnt<mdlhndl)) {
		// printf("Error [%s:%d]: invalid module handle\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL_ERR; break;
	}
{
	_psec_module_t psmp = _PSEC_STATE.pss_mdlslist[mdlhndl-1];
	if (!psmp||(mdlhndl!=((0xfff000&psmp->psm_stindex)>>16))) {
		// this should not happen, something went very wrong
		// printf("Error [%s:%d]: internal failure: no auth module\n", __FILE__, __LINE__);
		rc = PSEC_MODULE_HNDL2_ERR; break;
	}
	// check the module's state
	if (!(PSM_STATE_INITED&psmp->psm_stindex)) {
		if (rc = _psec_load_auth_module(psmp)) {
			// printf("Error [%s:%d]: failed to load and init auth module (%s)\n", __FILE__, __LINE__, psmp->psm_fpath);
			break;
		}
	}
	// check whether routine is supported by authentication module
	if (!psmp->psm_symtble[PSM__FREE_SIGNATURE]) {
		// printf("Error [%s:%d]: function not supported\n", __FILE__, __LINE__);
		rc = PSEC_NOT_SUPPORTED_ERR; break;
	}
	// call the module's routine
	if (rc = ((int(*)(psec_idbuf_t))psmp->psm_symtble[PSM__FREE_SIGNATURE])(signature)) {
		// printf("Error [%s:%d]: auth module's free_id_token() failed w/rc = %d\n", __FILE__, __LINE__, rc);
		rc = PSEC_MODULE_INTERNAL_ERR; break;
	}
}
} while (0);
	return rc;
}

// EOF
