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
#include <errno.h>
#include <stdio.h>
#include <string.h>
#include <assert.h>
#include <unistd.h>
#include <pwd.h>
#include <ctype.h>
#include <dlfcn.h>
#include <fcntl.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/param.h>

#include <openssl/pem.h>
#include <openssl/bio.h>
#include <openssl/md5.h>
#include <openssl/aes.h>

#include "psec_openssh.h"

#define USR_CONFIG_FILE ".ssh/config"
#define SSH_CONFIG_FILE "/etc/ssh/ssh_config"
#define SSHD_CONFIG_FILE "/etc/ssh/sshd_config"

#define CONFIG_FILE_MAXSIZE 10*1024

pthread_key_t _prngKey;
time_t _idtokTTL = 600;
char *authzkeyfile = NULL;
char *osslversion = NULL;
#if defined(_AIX)
#define _BIO_f_base64 BIO_f_base64
#define _BIO_free BIO_free
#define _BIO_free_all BIO_free_all
#define _BIO_new BIO_new
#define _BIO_new_mem_buf BIO_new_mem_buf
#define _BIO_push BIO_push
#define _BIO_read BIO_read
#define _BIO_set_flags BIO_set_flags
#define _BN_bin2bn BN_bin2bn
#define _BN_dec2bn BN_dec2bn
#define _BN_free BN_free
#define _DSA_free DSA_free
#define _DSA_new DSA_new
#define _DSA_sign DSA_sign
#define _DSA_size DSA_size
#define _DSA_verify DSA_verify
#define _PEM_read_DSAPrivateKey PEM_read_DSAPrivateKey
#define _PEM_read_RSAPrivateKey PEM_read_RSAPrivateKey
#define _RSA_free RSA_free
#define _RSA_new RSA_new
#define _RSA_sign RSA_sign
#define _RSA_size RSA_size
#define _RSA_verify RSA_verify
#define _RSA_public_encrypt RSA_public_encrypt
#define _RSA_private_decrypt RSA_private_decrypt
#define _SHA1 SHA1
#define _MD5_Init MD5_Init
#define _MD5_Update MD5_Update
#define _MD5_Final MD5_Final
#define _AES_set_encrypt_key AES_set_encrypt_key
#define _AES_encrypt AES_encrypt
#define _BIO_set_flags BIO_set_flags
#define OSSL_FNCSELECT(name) name
#elif defined(_LINUX)
typedef BIO_METHOD *(*_ft_BIO_f_base64)(void);
typedef int (*_ft_BIO_free)(BIO *);
typedef void (*_ft_BIO_free_all)(BIO *);
typedef BIO *(*_ft_BIO_new)(BIO_METHOD *);
typedef BIO *(*_ft_BIO_new_mem_buf)(void *, int);
typedef BIO *(*_ft_BIO_push)(BIO *, BIO *);
typedef int (*_ft_BIO_read)(BIO *, void *, int);
typedef int (*_ft_BN_dec2bn)(BIGNUM **, const char *);
typedef BIGNUM *(*_ft_BN_bin2bn)(const unsigned char *,int , BIGNUM *);
typedef void (*_ft_BN_free)(BIGNUM *);
typedef void (*_ft_DSA_free)(DSA *);
typedef DSA *(*_ft_DSA_new)(void);
typedef int (*_ft_DSA_sign)(int, const unsigned char *, int, unsigned char *, unsigned int *, DSA *);
typedef int (*_ft_DSA_size)(const DSA *);
typedef int (*_ft_DSA_verify)(int, const unsigned char *, int, const unsigned char *, int, DSA *);
typedef DSA *(*_ft_PEM_read_DSAPrivateKey)(FILE *, ...);
typedef RSA *(*_ft_PEM_read_RSAPrivateKey)(FILE *, ...);
typedef void (*_ft_RSA_free)(RSA *);
typedef RSA *(*_ft_RSA_new)(void);
typedef int (*_ft_RSA_sign)(int, const unsigned char *, unsigned int, unsigned char *, unsigned int *, RSA *);
typedef int (*_ft_RSA_size)(const RSA *);
typedef int (*_ft_RSA_verify)(int, const unsigned char *, unsigned int, unsigned char *, unsigned int, RSA *);
typedef int (*_ft_RSA_public_encrypt)(int, unsigned char *, unsigned char *, RSA *, int);
typedef int (*_ft_RSA_private_decrypt)(int, unsigned char *, unsigned char *, RSA *, int);
typedef unsigned char *(*_ft_SHA1)(const unsigned char *, size_t, unsigned char *);
typedef int (*_ft_MD5_Init)(MD5_CTX *);
typedef int (*_ft_MD5_Update)(MD5_CTX *, void *, size_t);
typedef int (*_ft_MD5_Final)(unsigned char *, MD5_CTX *);
typedef int (*_ft_AES_set_encrypt_key)(const unsigned char *, const int, AES_KEY *);
typedef void (*_ft_AES_encrypt)(const unsigned char *, unsigned char *, const AES_KEY *);
#if !defined(BIO_set_flags)
typedef void (*_ft_BIO_set_flags)(BIO *, int);
#endif

typedef struct func_doublet_s {
	char *fncname;
	void *fncpntr;
} func_doublet_desc, *func_doublet_t;

#define _fpi_BIO_f_base64 0
#define _fpi_BIO_free 1
#define _fpi_BIO_free_all 2
#define _fpi_BIO_new 3
#define _fpi_BIO_new_mem_buf 4
#define _fpi_BIO_push 5
#define _fpi_BIO_read 6
#define _fpi_BN_bin2bn 7
#define _fpi_BN_dec2bn 8
#define _fpi_BN_free 9
#define _fpi_DSA_free 10
#define _fpi_DSA_new 11
#define _fpi_DSA_sign 12
#define _fpi_DSA_size 13
#define _fpi_DSA_verify 14
#define _fpi_PEM_read_DSAPrivateKey 15
#define _fpi_PEM_read_RSAPrivateKey 16
#define _fpi_RSA_free 17
#define _fpi_RSA_new 18
#define _fpi_RSA_sign 19
#define _fpi_RSA_size 20
#define _fpi_RSA_verify 21
#define _fpi_RSA_public_encrypt 22
#define _fpi_RSA_private_decrypt 23
#define _fpi_SHA1 24
#define _fpi_MD5_Init 25
#define _fpi_MD5_Update 26
#define _fpi_MD5_Final 27
#define _fpi_AES_set_encrypt_key 28
#define _fpi_AES_encrypt 29
#if defined(BIO_set_flags)
#define OSSL_FNCSTBLE_SIZE 30
void *_fp_BIO_set_flags = NULL;
#else
#define _fpi_BIO_set_flags 30 
void _fp_BIO_set_flags(BIO *b, int flags) { b->flags |= flags; } 
#define OSSL_FNCSTBLE_SIZE 31
#endif
func_doublet_desc ossl_fncstble[OSSL_FNCSTBLE_SIZE] = {
	{"BIO_f_base64", NULL},
	{"BIO_free", NULL},
	{"BIO_free_all", NULL},
	{"BIO_new", NULL},
	{"BIO_new_mem_buf", NULL},
	{"BIO_push", NULL},
	{"BIO_read", NULL},
	{"BN_bin2bn", NULL},
	{"BN_dec2bn", NULL},
	{"BN_free", NULL},
	{"DSA_free", NULL},
	{"DSA_new", NULL},
	{"DSA_sign", NULL},
	{"DSA_size", NULL},
	{"DSA_verify", NULL},
	{"PEM_read_DSAPrivateKey", NULL},
	{"PEM_read_RSAPrivateKey", NULL},
	{"RSA_free", NULL},
	{"RSA_new", NULL},
	{"RSA_sign", NULL},
	{"RSA_size", NULL},
	{"RSA_verify", NULL},
	{"RSA_public_encrypt", NULL},
	{"RSA_private_decrypt", NULL},
	{"SHA1", NULL},
	{"MD5_Init", NULL},
	{"MD5_Update", NULL},
	{"MD5_Final", NULL},
	{"AES_set_encrypt_key", NULL},
	{"AES_encrypt", NULL}
#if !defined(BIO_set_flags)
	, {"BIO_set_flags", NULL}
#endif
};
#define OSSL_FNCSELECT(name) ((_ft##name)(ossl_fncstble[_fpi##name].fncpntr))
#endif

int
_read_config_param(
	char *cfgfile,
	char *param,
	char *value)
{
	int rc = 0;

do {
	struct stat cfgstat = {0};
	if (0>stat(cfgfile, &cfgstat)) {
		// printf("Error [%s:%d]: stat() failed for %s: errno = %d\n", __FILE__, __LINE__, cfgfile, errno);
		rc = OSSH_CFGFILE_ERR; break;
	}
	if (0==cfgstat.st_size) { rc = OSSH_CFGFILE_SIZE_ERR; break; }
	if (CONFIG_FILE_MAXSIZE<cfgstat.st_size) { rc = OSSH_CFGFILE_SIZE_ERR; break; }
{
	FILE *cfgstrm = NULL;
	if (NULL==(cfgstrm = fopen(cfgfile, "r"))) {
		// printf("Error [%s:%d]: fopen() failed for %s: errno = %d\n", __FILE__, __LINE__, cfgfile, errno);
		rc = OSSH_CFGFILE_OPEN_ERR; break;
 	}
	pthread_cleanup_push((void(*)(void *))fclose, (void *)cfgstrm);
do {
	char *cfgline = malloc(cfgstat.st_size);
	if (!cfgline) {
		// printf("Error [%s:%d]: fopen() failed w/errno = %d (%d)\n", __FILE__, __LINE__, errno, cfgstat.st_size);
		rc = OSSH_MEMORY_ERR; break;
	}
	pthread_cleanup_push((void(*)(void *))free, cfgline);
	do {
		char *cp = NULL;
		if (NULL==(cp = fgets(cfgline, cfgstat.st_size, cfgstrm))) {
			// if (ferror(cfgstrm)) printf("Error [%s:%d]: fgets() failed for %s: errno = %d\n", __FILE__, __LINE__, cfgfile, errno);
			if (feof(cfgstrm)) break;
			rc = OSSH_CFGFILE_GETS_ERR; break;
		}
		if (('#' == *cp) || ('\n' == *cp)) continue;
{
		while (isblank(*cp)) cp++;
		if (!strncmp(cp, param, strlen(param))) {
			char *pcp = cp + strlen(param);
			if (!isblank(*pcp)) continue;
			while (isblank(*pcp)) pcp++;
{
			char *ecp = pcp;
			while (ispunct(*ecp)||isalnum(*ecp)) ecp++;
			if (ecp == pcp) { rc = OSSH_CFGFILE_DATA_ERR; break; }
			if (PATH_MAX-1<ecp-pcp) { rc = OSSH_CFGFILE_DATA_ERR; break; }
			strncpy(value, pcp, ecp-pcp); value[ecp-pcp] = '\0';
			break;
}		
		}
}
	} while (1);
	pthread_cleanup_pop(1);		// free(cfgline);
} while (0);
	pthread_cleanup_pop(1);		// fclose(cfgstrm);
}
} while (0);
	return rc;
}

void
_nfree(void *p) { if (p) free(p); }

int
_get_identity_fname(
	char *luser,
	char *ruser,
	char *rhost,
	char **idfpath)
{
	// ruser and rhost not used at this time
	int rc = 0;
	char *usrConfigFile = NULL, *usrHomeDir = NULL;

	pthread_cleanup_push((void(*)(void *))_nfree, usrConfigFile);
	pthread_cleanup_push((void(*)(void *))_nfree, usrHomeDir);	
do {
	char vIdentityFile[PATH_MAX] = "";
	size_t usrHomeDirLen = 0;
	// get the user's home directory
	long pwrbufsize = sysconf(_SC_GETPW_R_SIZE_MAX);
	void *pwrbuf = malloc(pwrbufsize);
	if (!pwrbuf) {
		// printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		rc = OSSH_MEMORY_ERR; break;
	}
	pthread_cleanup_push((void(*)(void *))free, pwrbuf);
do {
	struct passwd usrpwd, *usrpwdp = NULL;
	if (luser&&('\0'!=*luser)) rc = getpwnam_r(luser, &usrpwd, pwrbuf, pwrbufsize, &usrpwdp);
	else rc = getpwuid_r(geteuid(), &usrpwd, pwrbuf, pwrbufsize, &usrpwdp);
	if (rc) {
		// printf("Error [%s:%d]: getpwuid_r() failed: rc = %d\n", __FILE__, __LINE__, rc);
		rc = OSSH_PWUID_ERR; break;
	}
	if (usrpwd.pw_dir&&(usrHomeDirLen = strlen(usrpwd.pw_dir))) {
		if(!(usrHomeDir = strdup(usrpwd.pw_dir))) {
			// printf("Error [%s:%d]: strdup() failed w/errno = %d\n", __FILE__, __LINE__, errno);
			rc = OSSH_MEMORY_ERR; break;
		}
		if (!(usrConfigFile = malloc(usrHomeDirLen+strlen(USR_CONFIG_FILE)+2))) {
			// printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
			rc = OSSH_MEMORY_ERR; break;
		}
		sprintf(usrConfigFile, "%s/%s", usrHomeDir, USR_CONFIG_FILE);
	}
} while (0);
	pthread_cleanup_pop(1);		// free(pwrbuf);
	if (rc) break;
do {
	if (!usrConfigFile) { rc = OSSH_CFGFILE_ERR; break; }
	rc = _read_config_param(usrConfigFile, "IdentityFile", vIdentityFile);
} while (0);
	if (rc||('\0'==*vIdentityFile)) {
		rc = _read_config_param(SSH_CONFIG_FILE, "IdentityFile", vIdentityFile);
	}
	if ((!rc)&&('\0' != *vIdentityFile)) {
		// check the file name
		if (strchr(vIdentityFile, '%')) {
			// printf("Error [%s:%d]: invalid format of identity file name: %s\n", __FILE__, __LINE__, vIdentityFile);
			rc = OSSH_IDFILE_NAME_ERR; break;
		}
do {
		if ('/' == *vIdentityFile) {
			// absolute file path
			rc = OSSH_IDFILE_NAME_ERR; break;
		}
		// relative file path to user's home directory
		if (!usrHomeDir) { rc = OSSH_PWUID_HOMEDIR_ERR; break; }
		if ('~' == *vIdentityFile) {
			if ('/' != *(vIdentityFile+1)) {
				// printf("Error [%s:%d]: invalid format of identity file name %s\n", __FILE__, __LINE__, vIdentityFile);
				rc = OSSH_IDFILE_NAME_ERR; break;
			}
			if (PATH_MAX-1<(usrHomeDirLen+strlen(vIdentityFile+1))) {
				// id file name too long
				rc = OSSH_IDFILE_PATH_ERR; break;
			}
			memmove(&vIdentityFile[usrHomeDirLen], vIdentityFile+1, strlen(vIdentityFile+1));
		} else {
			if (PATH_MAX-2<(usrHomeDirLen+strlen(vIdentityFile))) {
				// id file name too long
				rc = OSSH_IDFILE_PATH_ERR; break;
			}
			memmove(&vIdentityFile[usrHomeDirLen+1], vIdentityFile, strlen(vIdentityFile));
			vIdentityFile[usrHomeDirLen] = '/';
		}
		memcpy(vIdentityFile, usrHomeDir, usrHomeDirLen);
} while (0);
		if (rc) break;
{		// check the identity file
		struct stat idfstat = {0};
		if (0>stat(vIdentityFile, &idfstat)) {
			// printf("Error [%s:%d]: stat() failed for %s: errno = %d\n", __FILE__, __LINE__, idfpath, errno);
			rc = OSSH_IDFILE_PATH_ERR; break;
		}
		if (0==idfstat.st_size) {
			// printf("Error [%s:%d]: invalid private identity file (size = %d)\n", __FILE__, __LINE__, idfstat.st_size);
			rc = OSSH_IDFILE_SIZE_ERR; break;
		}
}
	} else {
		// either error reading config file or no identity file name
		// try the default file names: identity, id_rsa, id_dsa
		struct stat idfstat = {0};	
		strcpy(vIdentityFile, usrHomeDir);
do {
		rc = 0;		// reset error code, in case of any
		strcpy(vIdentityFile+usrHomeDirLen, "/.ssh/id_rsa");
		if ((0==stat(vIdentityFile, &idfstat))&&(0!=idfstat.st_size)) {
			// found this default file
			break;
		}
		memset(&idfstat, 0, sizeof(struct stat));
		strcpy(vIdentityFile+usrHomeDirLen, "/.ssh/id_dsa");
		if ((0==stat(vIdentityFile, &idfstat))&&(0!=idfstat.st_size)) {
			// found this default file
			break;
		}
		memset(&idfstat, 0, sizeof(struct stat));
		strcpy(vIdentityFile+usrHomeDirLen, "/.ssh/identity");
		if ((0==stat(vIdentityFile, &idfstat))&&(0!=idfstat.st_size)) {
			// found this default file
			break;
		}
		rc = OSSH_IDFILE_DEFS_ERR;
} while (0);
		if (rc) break;
	}
	if (!(*idfpath = strdup(vIdentityFile))) {
		// printf("Error [%s:%d]: strdup() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		rc = OSSH_MEMORY_ERR; break;
	}
} while (0);
	pthread_cleanup_pop(1);		// _nfree(usrHomeDir)
	pthread_cleanup_pop(1);		// _nfree(usrConfigFile)
	return rc;
}

int
_get_authz_fname(
	char *luser,
	char **azfpath)
{
	int rc = 0;
do {
	int dvAuthorizedFilef = 0, vAuthorizedFileLen = 0;
	char vAuthorizedFile[PATH_MAX] = "", *dvAuthorizedFilep = ".ssh/authorized_keys";
	if (!authzkeyfile) _read_config_param(SSHD_CONFIG_FILE, "AuthorizedKeysFile", vAuthorizedFile);
	else strcpy(vAuthorizedFile, authzkeyfile); 
do {
	if ('\0' == *vAuthorizedFile) {
		// no value found for authorized keys parameter
		strcpy(vAuthorizedFile, dvAuthorizedFilep);
		dvAuthorizedFilef++;
	}
	if ('/' == *vAuthorizedFile) {
		// absolute file path
		rc = OSSH_AUTHZFILE_NAME_ERR; break;
	}
	// relative path to user's home directory
{
	size_t usrHomeDirLen = 0;
	long pwrbufsize = sysconf(_SC_GETPW_R_SIZE_MAX);
	void *pwrbuf = malloc(pwrbufsize);
	if (!pwrbuf) {
		// printf("Error [%s:%d]: malloc() failed w/errno = %s\n", __FILE__, __LINE__, errno);
		rc = OSSH_MEMORY_ERR; break;
	}
	pthread_cleanup_push((void(*)(void *))free, pwrbuf);
do {
	struct passwd usrpwd, *usrpwdp = NULL;
	if (luser) rc = getpwnam_r(luser, &usrpwd, pwrbuf, pwrbufsize, &usrpwdp);
	else rc = getpwuid_r(geteuid(), &usrpwd, pwrbuf, pwrbufsize, &usrpwdp);
	if (rc) {
		// printf("Error [%s:%d]: getpwuid_r() failed: rc = %d\n", __FILE__, __LINE__, rc);
		rc = OSSH_PWUID_ERR; break;
	}
	if (!usrpwd.pw_dir|| (0==(usrHomeDirLen = strlen(usrpwd.pw_dir)))) {
		// printf("Error [%s:%d]: invalid user home directory\n", __FILE__, __LINE__);
		rc = OSSH_PWUID_HOMEDIR_ERR; break;
	}
	if ('~' == *vAuthorizedFile) {
		if ('/' != *(vAuthorizedFile+1)) {
			// printf("Error [%s:%d]: invalid format of authorized key file name: %s\n", __FILE__, __LINE__, vAuthorizedFile);
			rc = OSSH_AUTHZFILE_NAME_ERR; break;
		}
		if (PATH_MAX-1<(vAuthorizedFileLen = usrHomeDirLen+strlen(vAuthorizedFile+1))) {
			// authorized file name too long
			rc = OSSH_AUTHZFILE_PATH_ERR; break;
		}
		memmove(&vAuthorizedFile[usrHomeDirLen], vAuthorizedFile+1, strlen(vAuthorizedFile+1));
	} else {
		if (PATH_MAX-2<(vAuthorizedFileLen = usrHomeDirLen+strlen(vAuthorizedFile))) {
			// authorized file name too long
			rc = OSSH_AUTHZFILE_PATH_ERR; break;
		}
		memmove(&vAuthorizedFile[usrHomeDirLen+1], vAuthorizedFile, strlen(vAuthorizedFile));
		vAuthorizedFile[usrHomeDirLen] = '/';
	}
	memcpy(vAuthorizedFile, usrpwd.pw_dir, usrHomeDirLen);
} while (0);
	pthread_cleanup_pop(1);		// free(pwrbuf)
}
} while (0);
	if (rc) break;
{	// check the authorization file
	struct stat azfstat = {0};
	if (0>stat(vAuthorizedFile, &azfstat)) {
		if ((ENOENT!=errno)||(0==dvAuthorizedFilef)) {
			// printf("Error [%s:%d]: stat() failed w/errno = %d\n", __FILE__, __LINE__, errno);
			rc = OSSH_AUTHZFILE_PATH_ERR; break;
		}
		// default authorization file does not exist: try the 2 version
		memset(&azfstat, 0, sizeof(struct stat));
		if (PATH_MAX-2<vAuthorizedFileLen) {
			// too long
			rc = OSSH_AUTHZFILE_PATH_ERR; break;
		}
		strcat(vAuthorizedFile, "2");
		if (0>stat(vAuthorizedFile, &azfstat)) {
			// printf("Error [%s:%d]: stat() failed w/errno = %d\n", __FILE__, __LINE__, errno);
			rc = OSSH_AUTHZFILE_PATH_ERR; break;
		}
	}
}
	if (!(*azfpath = strdup(vAuthorizedFile))) {
		// printf("Error [%s:%d]: strdup() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		rc = OSSH_MEMORY_ERR; break;
	}
} while (0);
	return rc;
}

#define KEYTYPE_STR_DSA "ssh-dss"

DSA *
_read_dsa_public_key(
	char *s,
	char **usrid)
{
	DSA *dsapub = NULL;
do {
	int rc = 0;
	char *ts = strdup(s);
	if (!ts) {
		// printf("Error [%s:%d]: strdup() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		break;
	}
	pthread_cleanup_push((void(*)(void *))free, ts);
do {
	// find the dsa header in the public key data
	char *cp = strstr(ts, KEYTYPE_STR_DSA);
	if (!cp) {
		// not a DSA public key
		// printf("Error [%s:%d]: invalid DSA public key format\n", __FILE__, __LINE__);
		rc = OSSH_DSA_INVALID_FORMAT_ERR; break;
	}
{	// version 2 DSA public key
	char *lasts, *buf = NULL; size_t len;
	// get the base64 key encoding
	if (!(cp = strtok_r(cp+strlen(KEYTYPE_STR_DSA), " ", &lasts))) {
		// printf("Error [%s:%d]: invalid public key format\n", __FILE__, __LINE__);
		rc = OSSH_DSA_INVALID_FORMAT_ERR; break;
	}
	if (!(buf = malloc(2*(len=strlen(cp)+1)))) {
		// printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, len);
		rc = OSSH_MEMORY_ERR; break;
	}
	pthread_cleanup_push((void(*)(void *))free, buf);
do {
{	// decode the base64 encoded public key data
	BIO *bio, *b64;
	if (!(bio = OSSL_FNCSELECT(_BIO_new_mem_buf)(cp, -1))) {
		// printf("Error [%s:%d]: BIO_new_mem_buf() failed...\n", __FILE__, __LINE__);
		rc = OSSH_BIO_NEW_MEMBUF_ERR; break;
	}
	pthread_cleanup_push((void(*)(void *))OSSL_FNCSELECT(_BIO_free_all), bio);
do {
	if (!(b64 = OSSL_FNCSELECT(_BIO_new)(OSSL_FNCSELECT(_BIO_f_base64)()))) {
		// printf("Error [%s:%d]: BIO_new failed...\n", __FILE__, __LINE__);
		rc = OSSH_BIO_NEW_ERR; break;
	}
#if defined(BIO_set_flags)
	BIO_set_flags(b64, BIO_FLAGS_BASE64_NO_NL);
#else
	OSSL_FNCSELECT(_BIO_set_flags)(b64, BIO_FLAGS_BASE64_NO_NL);
#endif
	if (!(bio = OSSL_FNCSELECT(_BIO_push)(b64, bio))) {
		// printf("Error [%s:%d]: BIO_push() failed...\n", __FILE__, __LINE__);
		OSSL_FNCSELECT(_BIO_free)(b64);
		rc = OSSH_BIO_PUSH_ERR; break;
	}
	if (0>=(len = OSSL_FNCSELECT(_BIO_read)(bio, buf, len))) {
		// printf("Error [%s:%d]: BIO_read() failed\n", __FILE__, __LINE__);
		rc = OSSH_BIO_READ_ERR; break;
	}
} while (0);
	pthread_cleanup_pop(1);		// BIO_free_all(bio);
}
	if (rc) break;
{	// make sense of the public key data
	BIGNUM *p = NULL, *q = NULL, *g = NULL, *y= NULL;
	char *cp = buf;
	// 1st field should be key type
	int len = ntohl(*((int *)cp));
	if (strncmp(cp+=sizeof(int), KEYTYPE_STR_DSA, len)) {
		// printf("Error [%s:%d]: invalid public key format\n", __FILE__, __LINE__);
		rc = OSSH_DSA_INVALID_FORMAT_ERR; break;
	} else cp += len;
	// the subsequent fields for DSA should be the p, q, g and y parameters
	len = ntohl(*((int *)cp)); cp += sizeof(int);
do {
	// read the p parameter 
	if (!(p = OSSL_FNCSELECT(_BN_bin2bn)(cp, len, NULL))) {
		// printf("Error [%s:%d]: BN_bin2bn() failed for the p parameter...\n", __FILE__, __LINE__);
		rc = OSSH_BN_BIN2BN_ERR; break;
	} else cp += len;
	len = ntohl(*((int *)cp)); cp += sizeof(int);
	// read the q parameter
	if (!(q = OSSL_FNCSELECT(_BN_bin2bn)(cp, len, NULL))) {
		// printf("Error [%s:%d]: BN_bin2bn() failed for q parameter...\n", __FILE__, __LINE__);
		rc = OSSH_BN_BIN2BN_ERR; break;
	} else cp += len;
	len = ntohl(*((int *)cp)); cp += sizeof(int);
	// read the g parameter
	if (!(g = OSSL_FNCSELECT(_BN_bin2bn)(cp, len, NULL))) {
		// printf("Error [%s:%d]: BN_bin2bn() failed for the g parameter...\n", __FILE__, __LINE__);
		rc = OSSH_BN_BIN2BN_ERR; break;
	} else cp += len;
	len = ntohl(*((int *)cp)); cp += sizeof(int);
	// read the y parameter
	if (!(y = OSSL_FNCSELECT(_BN_bin2bn)(cp, len, NULL))) {
		// printf("Error [%s:%d]: BN_bin2bn() failed for the y parameter...\n", __FILE__, __LINE__);
		rc = OSSH_BN_BIN2BN_ERR; break;
	}
	if (!(dsapub = OSSL_FNCSELECT(_DSA_new)())) {
		// printf("Error [%s:%d]: DSA_new() failed...\n", __FILE__, __LINE__);
		rc = OSSH_DSA_NEW_ERR; break;
	}
	dsapub->p = p; dsapub->q = q; dsapub->g = g; dsapub->pub_key = y;
} while (0);
	if (rc) {
		if (p) OSSL_FNCSELECT(_BN_free)(p);
		if (q) OSSL_FNCSELECT(_BN_free)(q);
		if (g) OSSL_FNCSELECT(_BN_free)(g);
		if (y) OSSL_FNCSELECT(_BN_free)(y);
	}
}
} while (0);
	pthread_cleanup_pop(1);		// free(buf)
	if (rc) break;
	if (usrid&&(cp = strtok_r(NULL, " \n", &lasts))) *usrid = strdup(cp);
}
} while (0);
	pthread_cleanup_pop(1);		// free (ts);
} while (0);
	return dsapub;
}

#define KEYTYPE_STR_RSA "ssh-rsa"

RSA *
_read_rsa_public_key(
	char *s,
	char **usrid)
{
	// this routine reads a public RSA key from either an openSSH v1 or v2
	// formatted public key identity file
	// the v1 format appears to be as follows: options, key length, public
	// exponent, modulus, comments; the fields are separated by white
	// spaces
	// the v2 format appears to be as follows: options, key type, key
	// value, comment; the fields are separated by one blank space
	// in both versions options are optional; the comment field is not used
	RSA *rsapub = NULL;
do {
	int rc = 0;
	char *ts = strdup(s);
	if (!ts) {
		// printf("Error [%s:%d]: strdup() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		break;
	}
	pthread_cleanup_push((void(*)(void *))free, ts);
do {
	// if rsa header, then RSA v2 key
	char *cp = strstr(ts, KEYTYPE_STR_RSA);
	if (!cp) {
		// possibly a v1 RSA public key 	
		// split the input string in tokens separated by blank spaces
		int noofquotes = 0; char *lasts, *cp = strtok_r(ts, " ", &lasts); 
		if (!cp) {
			// printf("Error [%s:%d]: invalid public key format\n", __FILE__, __LINE__);
			rc = OSSH_RSA_INVALID_FORMAT_ERR; break;
		}
		if (!isdigit(*cp)) {
			// this must be a comment
			int quotesclosed = 0;
			while (!quotesclosed) {
{				// check whether a blank space whithin quotes
				char *cp2 = cp-1;
				while (cp2 = strchr(cp2+1, '"')) noofquotes++;
				if (!(noofquotes%2)) quotesclosed++;
				if (!(cp = strtok_r(NULL, " ", &lasts))) {
					// printf("Error [%s:%d]: invalid public key format\n", __FILE__, __LINE__);
					rc = OSSH_RSA_INVALID_FORMAT_ERR; break;
				} 
}
			}
			if (!isdigit(*cp)) {
				// printf("Error [%s:%d]: invalid public key format\n", __FILE__, __LINE__);
				rc = OSSH_RSA_INVALID_FORMAT_ERR; break;
			}
		}
		// version 1 public key; get the key length
		BIGNUM *e = NULL, *n = NULL;
		long keylen;
		if ((0==(errno = 0, keylen=atol(cp)))&&(EINVAL==errno)) {
			// printf("Error [%s:%d]: invalid public key format\n", __FILE__, __LINE__);
			rc = OSSH_RSA_INVALID_FORMAT_ERR; break;
		}
		if (((LONG_MAX==keylen)||(LONG_MIN==keylen))&&(ERANGE==errno)) {
			// printf("Error [%s:%d]: invalid public key format\n", __FILE__, __LINE__);
			rc = OSSH_RSA_INVALID_FORMAT_ERR; break;
		}
		if (0>=keylen) {
			// printf("Error [%s:%d]: invalid public key format\n", __FILE__, __LINE__);
			rc = OSSH_RSA_INVALID_FORMAT_ERR; break;
		}
		// get the public exponent
		if (!(cp = strtok_r(NULL, " ", &lasts))) {
			// printf("Error [%s:%d]: invalid public key format\n", __FILE__, __LINE__);
			rc = OSSH_RSA_INVALID_FORMAT_ERR; break;
		}
		if (!(OSSL_FNCSELECT(_BN_dec2bn)(&e, cp))) {
			// printf("Error [%s:%d]: BN_dec2bn() failed...\n", __FILE__, __LINE__);
			rc = OSSH_BN_DEC2BN_ERR; break;
		}
do {
		// get the modulus
		if (!(cp = strtok_r(NULL, " ", &lasts))) {
			// printf("Error [%s:%d]: invalid public key format\n", __FILE__, __LINE__);
			rc = OSSH_RSA_INVALID_FORMAT_ERR; break;
		}
		if (!(OSSL_FNCSELECT(_BN_dec2bn)(&n, cp))) {
			// printf("Error [%s:%d]: BN_dec2bn() failed...\n", __FILE__, __LINE__);
			rc = OSSH_BN_DEC2BN_ERR; break;
		}
do {
		if (!(rsapub = OSSL_FNCSELECT(_RSA_new)())) {
			// printf("Error [%s:%d]: RSA_new() failed...\n", __FILE__, __LINE__);
			rc = OSSH_RSA_NEW_ERR; break;
		}
		rsapub->n = n; rsapub->e = e;
} while (0);
		if (rc) { OSSL_FNCSELECT(_BN_free)(n); break; }
} while (0);
		if (rc) { OSSL_FNCSELECT(_BN_free)(e); break; }
		// get the user id, if any--normally, there should be a
		// user@hostname id
		if (usrid&&(cp = strtok_r(NULL, " \n", &lasts))) *usrid = strdup(cp);
	} else {
		// version 2 public key
		char *lasts, *buf = NULL; size_t len;
		// get the base64 key encoding
		if (!(cp = strtok_r(cp+strlen(KEYTYPE_STR_RSA), " ", &lasts))) {
			// printf("Error [%s:%d]: invalid RSA public key format\n", __FILE__, __LINE__);
			rc = OSSH_RSA_INVALID_FORMAT_ERR; break;
		}
		if (!(buf = malloc(2*(len=strlen(cp)+1)))) {
			// printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, len);
			rc = OSSH_MEMORY_ERR; break;
		}
		pthread_cleanup_push((void(*)(void *))free, buf);
do {
{		// decode the base64 encoded public key data
		BIO *bio, *b64;
		if (!(bio = OSSL_FNCSELECT(_BIO_new_mem_buf)(cp, -1))) {
			// printf("Error [%s:%d]: BIO_new_mem_buf() failed...\n", __FILE__, __LINE__);
			rc = OSSH_BIO_NEW_MEMBUF_ERR; break;
		}
		pthread_cleanup_push((void(*)(void *))OSSL_FNCSELECT(_BIO_free_all), bio);
do {
		if (!(b64 = OSSL_FNCSELECT(_BIO_new)(OSSL_FNCSELECT(_BIO_f_base64)()))) {
			// printf("Error [%s:%d]: BIO_new failed...\n", __FILE__, __LINE__);
			rc = OSSH_BIO_NEW_ERR; break;
		}
#if defined(BIO_set_flags)
		BIO_set_flags(b64, BIO_FLAGS_BASE64_NO_NL);
#else
		OSSL_FNCSELECT(_BIO_set_flags)(b64, BIO_FLAGS_BASE64_NO_NL);
#endif
		if (!(bio = OSSL_FNCSELECT(_BIO_push)(b64, bio))) {
			// printf("Error [%s:%d]: BIO_push() failed...\n", __FILE__, __LINE__);
			OSSL_FNCSELECT(_BIO_free)(b64);
			rc = OSSH_BIO_PUSH_ERR; break;
		}
		if (0>=(len = OSSL_FNCSELECT(_BIO_read)(bio, buf, len))) {
			// printf("Error [%s:%d]: BIO_read() failed...\n", __FILE__, __LINE__);
			rc = OSSH_BIO_READ_ERR; break;
		}
} while (0);
		pthread_cleanup_pop(1);		// BIO_free_all(bio);
}
		if (rc) break;
{		// make sense of the public key data
		BIGNUM *e = NULL, *n = NULL;
		char *cp = buf;
		// check the key type
		int len = ntohl(*((int *)cp));
		if (strncmp(cp+=sizeof(int), KEYTYPE_STR_RSA, len)) {
			// printf("Error [%s:%d]: invalid public key format\n", __FILE__, __LINE__);
			rc = OSSH_RSA_INVALID_FORMAT_ERR; break;
		} else cp += len;
do {
		// read the public exponent
		len = ntohl(*((int *)cp)); cp += sizeof(int);
		if (!(e = OSSL_FNCSELECT(_BN_bin2bn)(cp, len, NULL))) {
			// printf("Error [%s:%d]: BN_bin2bn() failed for the public exponent...\n", __FILE__, __LINE__);
			rc = OSSH_BN_BIN2BN_ERR; break;
		}
		cp += len;
		// read the modulus
		len = ntohl(*((int *)cp));
		cp += sizeof(int);
		// read the modulus
		if (!(n = OSSL_FNCSELECT(_BN_bin2bn)(cp, len, NULL))) {
			// printf("Error [%s:%d]: BN_bin2bn() failed for modulus...\n", __FILE__, __LINE__);
			rc = OSSH_BN_BIN2BN_ERR; break;
		}
		if (!(rsapub = OSSL_FNCSELECT(_RSA_new)())) {
			// printf("Error [%s:%d]: RSA_new() failed...\n", __FILE__, __LINE__);
			rc = OSSH_RSA_NEW_ERR; break;
		}
		rsapub->n = n; rsapub->e = e;
} while (0);
		if (rc) {
			if (e) OSSL_FNCSELECT(_BN_free)(e);
			if (n) OSSL_FNCSELECT(_BN_free)(n);
		}
}
} while (0);
		pthread_cleanup_pop(1);		// free(buf);
		if (rc) break;
		if (usrid&&(cp = strtok_r(NULL, " \n", &lasts))) *usrid = strdup(cp);
	}
} while (0);
	pthread_cleanup_pop(1);		// free(ts);
} while (0);
	return rsapub;
}

#define SSHRSA1_FILE_IDSTR	"SSH PRIVATE KEY FILE FORMAT 1.1\n"
RSA *
_read_rsa1_private_key(
	FILE *strm)
{
	RSA *rsaprv = NULL;
do {
	// get the file desriptor number and check the size of the private key
	// file
	int strmfsize = 0; char *strmfbuf;
	int strmfd = fileno(strm);
	if (0>strmfd) {
		// printf("Error [%s:%d]: fileno() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		break;
	}
{
	struct stat strmstat = {0};
	if (0>fstat(strmfd, &strmstat)) {
		// printf("Error [%s:%d]: fstat() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		break;
	}
	if ((0==(strmfsize = strmstat.st_size))||(1024*1024<strmfsize)) {
		// printf("Error [%s:%d]: private key file out of range (%d bytes)\n", __FILE__, __LINE__, strmfsize);
		break;
	}
}
	// allocate memory for the private key file content
	if(!(strmfbuf = malloc(strmfsize))) {
		// printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		break;
	}
	pthread_cleanup_push((void(*)(void *))free, strmfbuf);
do {
	// read the private key file
	size_t nbytes;
	if (0>(nbytes = read(strmfd, strmfbuf, strmfsize))) {
		// printf("Error [%s:%d]: read() failed w/errno = %d\n", __FILE__, __LINE__, nbytes);
		break;
	}
	if (nbytes != strmfsize) {
		// printf("Error [%s:%d]: failed to read the entire private key file\n", __FILE__, __LINE__);
		break;
	}
{
	// check the private key file contents
	char *cp = strmfbuf;
	if (strcmp(SSHRSA1_FILE_IDSTR, cp)) {
		// printf("Error [%s:%d]: not an RSA1 private key file\n", __FILE__, __LINE__);
		break;
	}
	cp += strlen(SSHRSA1_FILE_IDSTR)+1;
	// check the cipher type
	if (0!=cp[0]) {
		// printf("Error [%s:%d]: no support for passphrase protected private key files\n", __FILE__, __LINE__);
		break;
	} else cp++;
	// skip the reserved data
	cp += sizeof(int);
{
	int rc = 0;
	BIGNUM *n = NULL, *e = NULL, *d = NULL, *iqmp = NULL, *p = NULL, *q = NULL;
do {
	// next is the public key
	int pkeysize; short bnsize; size_t bnbytes;
	// the key size -- not really used
	pkeysize = ntohl(*((unsigned int *)cp)); cp += sizeof(unsigned int);
	// the modulus
	bnsize = ntohs(*((unsigned short int *)cp)); cp += sizeof(unsigned short int);
	bnbytes = (bnsize +7)/8;
	if (!(n = OSSL_FNCSELECT(_BN_bin2bn)(cp, bnbytes, NULL))) {
		// printf("Error [%s:%d]: BN_bin2bn() failed for modulus\n", __FILE__, __LINE__);
		rc = OSSH_BN_BIN2BN_ERR; break;
	} else cp += bnbytes;
	// the public exponent
	bnsize = ntohs(*((unsigned short int *)cp)); cp += sizeof(unsigned short int);
	bnbytes = (bnsize+7)/8;
	if (!(e = OSSL_FNCSELECT(_BN_bin2bn)(cp, bnbytes, NULL))) {
		// printf("Error [%s:%d]: BN_bin2bn() failed for public exponent\n", __FILE__, __LINE__);
		rc = OSSH_BN_BIN2BN_ERR; break;
	} else cp += bnbytes;
	// skip the comment
	bnbytes = ntohl(*((unsigned int *)cp)); cp += sizeof(unsigned int);
	cp += bnbytes;
	// check the cipher verification data
	if ((cp[0]!=cp[2])||(cp[1]!=cp[3])) {
		// printf("Error [%s:%d]: cipher verification failure\n", __FILE__, __LINE__);
		rc = OSSH_CIPHER_VERIF_ERR; break;
	} else cp += sizeof(unsigned int);
	// get private key
	bnsize = ntohs(*((unsigned short int *)cp)); cp += sizeof(unsigned short int);
	bnbytes = (bnsize+7)/8;
	if (!(d = OSSL_FNCSELECT(_BN_bin2bn)(cp, bnbytes, NULL))) {
		// printf("Error [%s:%d]: BN_bin2bn() failed for private key\n", __FILE__, __LINE__);
		rc = OSSH_BN_BIN2BN_ERR; break;
	} else cp += bnbytes;
	// get (q^-1)mod(p)
	bnsize = ntohs(*((unsigned short int *)cp)); cp += sizeof(unsigned short int);
	bnbytes = (bnsize+7)/8;
	if (!(iqmp = OSSL_FNCSELECT(_BN_bin2bn)(cp, bnbytes, NULL))) {
		// printf("Error [%s:%d]: BN_bin2bn() failed for (q^-1)mod(p)\n", __FILE__, __LINE__);
		rc = OSSH_BN_BIN2BN_ERR; break;
	} else cp += bnbytes;
	// get prime factors
	bnsize = ntohs(*((unsigned short int *)cp)); cp += sizeof(unsigned short int);
	bnbytes = (bnsize+7)/8;
	if (!(q = OSSL_FNCSELECT(_BN_bin2bn)(cp, bnbytes, NULL))) {
		// printf("Error [%s:%d]: BN_bin2bn() failed for prime factor q\n", __FILE__, __LINE__);
		rc = OSSH_BN_BIN2BN_ERR; break;
	} else cp += bnbytes;
	bnsize = ntohs(*((unsigned short int *)cp)); cp += sizeof(unsigned short int);
	bnbytes = (bnsize+7)/8;
	if (!(p = OSSL_FNCSELECT(_BN_bin2bn)(cp, bnbytes, NULL))) {
		// printf("Error [%s:%d]: BN_bin2bn() failed for prime factor p\n", __FILE__, __LINE__);
		rc = OSSH_BN_BIN2BN_ERR; break;
	}
	// create an RSA key and fill it in
	if (!(rsaprv = OSSL_FNCSELECT(_RSA_new)())) {
		// printf("Error [%s:%d]: RSA_new() failed...\n", __FILE__, __LINE__);
		rc = OSSH_RSA_NEW_ERR; break;
	}
	rsaprv->n = n; rsaprv->e = e; rsaprv->d = d;
	rsaprv->p = p; rsaprv->q = q; rsaprv->iqmp = iqmp;
} while (0);
	if (rc) {
		if (n) OSSL_FNCSELECT(_BN_free)(n);
		if (e) OSSL_FNCSELECT(_BN_free)(e);
		if (d) OSSL_FNCSELECT(_BN_free)(d);
		if (iqmp) OSSL_FNCSELECT(_BN_free)(iqmp);
		if (p) OSSL_FNCSELECT(_BN_free)(p);
		if (q) OSSL_FNCSELECT(_BN_free)(q);
	}
}
}
} while (0);
	pthread_cleanup_pop(1);		// free(strmfbuf);
} while (0);
	return rsaprv;
}

#include <time.h>

int
_prng_random(unsigned int *prn)
{
	int rc = 0;
	unsigned int *_prngData;
do {
	// get the thread specific data for the PRNG
	if (!(_prngData = pthread_getspecific(_prngKey))) {
		// allocate memory for the thread specific PRNG data
		if (!(_prngData = calloc(1, sizeof(unsigned int)))) {
			// printf("Error [%s:%d]: calloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
			rc = OSSH_MEMORY_ERR; break;
		}
{		// seed the PRNG data
#ifdef _AIX
		struct timespec ts = {0};
		clock_gettime(CLOCK_REALTIME, &ts);
		*_prngData = (unsigned int)(ts.tv_sec^ts.tv_nsec);
#else
		struct timeval tv = {0};
		gettimeofday(&tv, NULL);
		*_prngData = (unsigned int)(tv.tv_sec^tv.tv_usec);
#endif
}
		// set the PRNG data
		if (rc = pthread_setspecific(_prngKey, _prngData)) {
			// printf("Error [%s:%d]: pthread_setspecific() failed w/rc = %d\n", __FILE__, __LINE__, rc);
			rc = OSSH_PTHRD_SETSPECIFIC_ERR; break;
		}
	}
{	// mangle the _prngData
#ifdef _AIX
	struct timespec ts = {0};
	clock_gettime(CLOCK_REALTIME, &ts);
	*_prngData |= (unsigned int)(ts.tv_sec^ts.tv_nsec);
#else
	struct timeval tv = {0};
	gettimeofday(&tv, NULL);
	*_prngData |= (unsigned int)(tv.tv_sec^tv.tv_usec);
#endif
}
	*prn = rand_r(_prngData); 
} while (0);
	return rc;
}

int
_generate_session_key(
	RSA *pub,
	unsigned char *key,
	size_t *keylen)
{
	// this routine generates a 128bit AES key and encrypts it with the
	// RSA public key provided
	// will not check the arguments because I control how this routine
	// is called
	int rc = 0;
	unsigned char rawkey[16]; size_t tkeylen;
do {
	// generate the key
	int i = 0; for(;i<4; i++) if (rc = _prng_random((unsigned int *)((void *)rawkey+i*sizeof(unsigned int)))) break;
	if (rc) break;
	// encrypt the key
	if (0>(tkeylen = OSSL_FNCSELECT(_RSA_public_encrypt)(16, rawkey, key, pub, RSA_PKCS1_PADDING))) {
		// printf("Error [%s:%d]: RSA_public_encrypt() failed for session key\n", __FILE__, __LINE__);
		rc = OSSH_RSA_PUBLIC_ENCRYPT_ERR; break;
	}
	*keylen = tkeylen;
} while (0);
	return rc;
}

int
_recover_session_key(
	RSA *prv,
	unsigned char *ekey,
	size_t ekeylen,
	unsigned char *key,
	size_t *keylen)
{
	// this routine recovers the key data from the encrypted key
	// material in the identity token; the encrypted material was
	// generated by encrypting a raw 128bit AES key with the user's
	// OpenSSH RSA public key; therefore, the recovery of the session
	// key uses the user's OpenSSH RSA private key to decrypt the
	// encrypted key material
	// it assumes valid arguments
	int rc = 0;
	unsigned char *tbuf;
do {
	if (16>*keylen) {
		// not enough memory to store the return session key
		// printf("Error [%s:%d]: not enough memory for the session key; bytes required: 16\n", __FILE__, __LINE__):
		*keylen=16; rc = OSSH_MEMORY_ERR; break;
	}
	// allocate memory for a temporary buffer
	if (!(tbuf = malloc(OSSL_FNCSELECT(_RSA_size)(prv)))) {
		// unable to allocate memory for the encrypted key
		// printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		*keylen = 0; rc = OSSH_MEMORY_ERR; break;
	}
do {
	// decrypt the encrypted key
	if (16 != OSSL_FNCSELECT(_RSA_private_decrypt)(ekeylen, ekey, tbuf, prv, RSA_PKCS1_PADDING)) {
		rc = OSSH_RSA_PRIVATE_DECRYPT_ERR; break;
	}
	memcpy(key, tbuf, *keylen=16); 
} while (0);
	free(tbuf);
} while (0);
	return rc;
}

int
psm__init(
	char *opts)
{
	int rc = 0;
do {
	// create a pthread key for the pseudo-random specific data
	if (rc = pthread_key_create(&_prngKey, free)) {
		// printf("Error [%s:%d]: pthread_key_create() failed w/rc = %d\n", __FILE__, __LINE__, rc);
		rc = OSSH_PTHRD_KEYCREATE_ERR; break;
	}
	if (!opts||('\0'==opts)) break;
{	char *tcp, *lasts, *cp = strdup(opts);
	if (!cp) {
		// printf("Error [%s:%d]: strdup() failed w/errno = %s\n", __FILE__, __LINE__, errno);
		rc = OSSH_MEMORY_ERR; break;
	}
do {
	tcp = strtok_r(cp, ",", &lasts);
	do {
		switch (*tcp) {
			case 't':
				if ('='!=*++tcp) break;
{				time_t tidtokTTL; if (0==(tidtokTTL = strtol(++tcp, NULL, 10))) {
					// printf("Error [%s:%d]: TTL value must be a positive number\n", __FILE__, __LINE__);
					break;
				}
				if (0>tidtokTTL) _idtokTTL = 0;
				else _idtokTTL = tidtokTTL;
}
				break;
			case 'v':
				if ('='!=*++tcp) break;
				if (PATH_MAX-25<strlen(++tcp)) {
					// printf("Error [%s:%d]: Version must have a valid format for openssl\n", __FILE__, __LINE__);
					break;
				}
				osslversion = strdup(tcp);
				break;
			case 'z':
				if ('='!=*++tcp) break;
				if (PATH_MAX-1<strlen(++tcp)) {
					// printf("Error [%s:%d]: Authorized keys file must be a valid file name\n", __FILE__, __LINE__);
					break;
				}
				authzkeyfile = strdup(tcp);
				break;
			default:
				// skip unknown options
				break;
		}
	} while (tcp = strtok_r(NULL, ",", &lasts));
} while (0);
	free(cp);
}
{	// open the libcrypto.so library and resolve required symbols
	char osslcryptolib[PATH_MAX+1] = "", *errmsg = NULL; int i; void *p;
#if defined(_LINUX)
#if defined(__64BIT__)
	strcpy(osslcryptolib, "/usr/lib64/libcrypto.so");
#else
	strcpy(osslcryptolib, "/usr/lib/libcrypto.so");
#endif
	if (osslversion) {
		strcat(osslcryptolib, ".");
		strcat(osslcryptolib, osslversion);
	}
	p = dlopen(osslcryptolib, RTLD_NOW);
	if (!p) {
		char *errmsg = dlerror();
		// printf("Error [%s:%d]: dlopen() failed: %s\n", __FILE__, __LINE__, errmsg?errmsg:"<no error message>");
		rc = OSSH_DLOPEN_ERR; break;
	}
	for (i=0; i<OSSL_FNCSTBLE_SIZE; i++) {
		if (!(ossl_fncstble[i].fncpntr = dlsym(p, ossl_fncstble[i].fncname))) {
			if (!strcmp(ossl_fncstble[i].fncname, "BIO_set_flags")) {
				ossl_fncstble[i].fncpntr = _fp_BIO_set_flags;
				continue;
			}
			errmsg = dlerror();
			// printf("Error [%s:%d]: dlsym() failed for %s: %s\n", __FILE__, __LINE__, ossl_fncstble[i].fncname, errmsg?errmsg:"<no error message>");
			rc = OSSH_DLSYM_ERR; break;
		}
	}
	if (rc) break;
#elif !defined(_AIX)
	rc = OSSH_DLOPEN_ERR; break;
#endif
}
} while (0);
	return rc;
}

void
psm__cleanup()
{
do {
	// nothing to do
} while (0);
	return;
}

typedef union {
	u_int64_t l;
	u_int32_t i[2];
	u_int16_t s[4];
	u_int8_t  c[8];
} Uu_int64_t;

#define IDTOK_LEN_MIN 32
#define IDTOK_LEN (4*1024)
int
_increase_tknsize(
	size_t len,
	void **p,
	size_t *nlen) 
{
	int rc = 0;
do {
	void *tp = NULL; size_t tlen = (len/IDTOK_LEN+1)*IDTOK_LEN;
	if (!(tp = realloc(*p, tlen))) {
		// printf("Error [%s:d]: realloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		rc = OSSH_MEMORY_ERR; break;
	}
	*p = tp; *nlen = tlen;
} while (0);
	return rc;
}

int
psm__get_id_token(
	char *ruser,
	char *rhost,
	psm_idbuf_t idtok)
{
	int rc = 0;
do {
	if (!idtok) {
		// printf("Error [%s:%d]: invalid id buffer descriptor\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	}
{
	char *idfpath = NULL, *pubidfpath = NULL;
	if (rc = _get_identity_fname(NULL, ruser, rhost, &idfpath)) break;
	pthread_cleanup_push((void(*)(void *))free, idfpath);
do {
	if (!(pubidfpath = malloc(strlen(idfpath)+strlen(".pub")+1))) {
		// printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		rc = OSSH_MEMORY_ERR; break;
	}
	pthread_cleanup_push((void(*)(void *))free, pubidfpath);
do {
	strcpy(pubidfpath, idfpath);
	strcat(pubidfpath, ".pub");
{
	struct stat idfstat = {0};
	if (0>stat(pubidfpath, &idfstat)) {
		// printf("Error [%s:%d]: stat() failed for %s: errno = %d\n", __FILE__, __LINE__, pubidfpath, errno);
		rc = OSSH_PIDFILE_PATH_ERR; break;
	}
	if (0==idfstat.st_size) {
		// printf("Error [%s:%d]: invalid public identity file (size = %d)\n", __FILE__, __LINE__, idfstat.st_size);
		rc = OSSH_PIDFILE_SIZE_ERR; break;
	}
}
{	// read the keys
	RSA *rsaprv = NULL, *rsapub = NULL;
	DSA *dsaprv = NULL, *dsapub = NULL;
	char usridstr[1024] = ""; int usridstrlen = strlen(usridstr);
do {
{	// open the private identity file
	FILE *idfstrm = fopen(idfpath, "r");
	if (!idfstrm) {
		// printf("Error [%s:%d]: fopen() failed for %s: errno = %d\n", __FILE__, __LINE__, idfpath, errno);
		rc = OSSH_IDFILE_OPEN_ERR; break;
	}
	pthread_cleanup_push((void(*)(void *))fclose, idfstrm);
do {
	// read the private key
	if (rsaprv = _read_rsa1_private_key(idfstrm)) break;
	// rewind the stream to the beginning
	fseek(idfstrm, 0, SEEK_SET);
	if (rsaprv = OSSL_FNCSELECT(_PEM_read_RSAPrivateKey)(idfstrm, NULL, NULL, "")) break;
	// rewind the stream to the beginning
	fseek(idfstrm, 0, SEEK_SET);
	if (dsaprv = OSSL_FNCSELECT(_PEM_read_DSAPrivateKey)(idfstrm, NULL, NULL, "")) break;
	// printf("Error [%s:%d]: unable to read user's private key\n", __FILE__, __LINE__);
	rc = OSSH_IDFILE_READ_ERR; break;
} while (0);
	pthread_cleanup_pop(1);		// fclose(idfstrm);
	if (rc) break;
}
	// the type of key in the public identity file can be derived from
	// the type of the private key; if the private key is a DSA key,
	// the public key must be also a DSA key; if the private key is an
	// RSA key, the public identity file may containe either an SSH-1
	// or an SSH-2 formatted RSA public key
{	// open the public identity file
	char *usrid = NULL; FILE *idfstrm = NULL;
	pthread_cleanup_push((void(*)(void *))_nfree, usrid);
do {
	if (!(idfstrm=fopen(pubidfpath, "r"))) {
		// printf("Error [%s:%d]: fopen() failed for %s: errno = %d\n", __FILE__, __LINE__, pubidfpath, errno);
		rc = OSSH_PIDFILE_OPEN_ERR; break;
	}
	pthread_cleanup_push((void(*)(void *))fclose, idfstrm);
do {
	char *cp, pubkeydata[8*1024] = "";
	if (!(cp = fgets(pubkeydata, 8*1024, idfstrm))) {
		// printf("Error [%s:%d]: fgets() failed or EOF\n", __FILE__, __LINE__);
		rc = OSSH_PIDFILE_GETS_ERR; break;
	}
	if (dsaprv) {
		// this must be a DSA public key file
		if (!(dsapub = _read_dsa_public_key(cp, &usrid))) rc = OSSH_DSA_INVALID_FORMAT_ERR;
		break;
	}
	if (!(rsapub = _read_rsa_public_key(cp, &usrid))) { rc = OSSH_RSA_INVALID_FORMAT_ERR; break; }
} while (0);
	pthread_cleanup_pop(1);		// fclose(idfstrm);
	if (rc) break;
{
	// build the id token and sign it; start w/a 4k buffer and realloc if
	// necessary
	// the id token's structure is as follows:
	//	4 bytes: magic - "ossh"
#define IDTOK_OFFST_STATVER 4
	//	1 byte: status (high nibble) and version (low nibble)
#define IDTOK_OFFST_KEYTYPE 5
	//	1 byte: key type - 1 for DSA, 2 for RSA
#define IDTOK_OFFST_NONCE 6
	//	8 bytes: nonce
#define IDTOK_OFFST_TMSTAMP 14
	//	8 bytes: time stamp
#define IDTOK_OFFST_IDDATA 22
	//	2 bytes: length of target user name - N bytes
	//	N bytes: target user name
	//	2 bytes: length of target host name - M bytes
	//	M bytes: target host name
	//	2 bytes: length of user name - P bytes
	//	P bytes: user name
	//  2 bytes: length of encrypted session key - Q bytes
	//  Q bytes: encrypted session key
	//	2 bytes: length of id token signature - R bytes
	//	R bytes: id token signature
	//	minimum length of the id token is 34 bytes, including a minimum
	//  1 byte for the user id and a minimum 1 byte for the signature
#define IDTOK_OFFST_MINLEN 32

#define IDTOK_MAGIC "ossh"
	size_t idtoklen = IDTOK_LEN, cidlen = 0;
	char *idtokbuf = (char *)malloc(IDTOK_LEN);
	if (!idtokbuf) {
		// printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		rc = OSSH_MEMORY_ERR; break;
	}
do {	// no cancellations points in this loop
	// write the magic, version and key type
	sprintf(idtokbuf, "%s", IDTOK_MAGIC); cidlen += strlen(IDTOK_MAGIC);
	*((unsigned char *)(idtokbuf+cidlen)) = 0x02; cidlen += sizeof(unsigned char);
	*((unsigned char *)(idtokbuf+cidlen)) = dsaprv?0x01:0x02; cidlen += sizeof(unsigned char);
{	// write the nonce...
	Uu_int64_t nonce;
	_prng_random((unsigned int *)&nonce.i[0]); _prng_random((unsigned int *)&nonce.i[1]);
	*((u_int32_t *)(idtokbuf+cidlen)) = htonl(nonce.i[0]); cidlen += sizeof(nonce.i[0]);
	*((u_int32_t *)(idtokbuf+cidlen)) = htonl(nonce.i[1]); cidlen += sizeof(nonce.i[0]);
	// ...and skip the time stamp--will do that later
	cidlen += sizeof(Uu_int64_t);
}
{	// write the remote user id
	unsigned short ruserlen = ruser?strlen(ruser):0;
	if (idtoklen<(cidlen+sizeof(unsigned long)+ruserlen)) {
		// not enough space in the allocated id token buffer--need to
		// reallocate
		if (rc = _increase_tknsize(cidlen+sizeof(unsigned long)+ruserlen, (void **)&idtokbuf, &idtoklen)) break;
	}
	// write the remote user id length... 
	*((unsigned short *)(idtokbuf+cidlen)) = htons(ruserlen); cidlen += sizeof(unsigned short);
	// ...and the remote user id
	if (ruserlen) { sprintf(idtokbuf+cidlen, "%s", ruser); cidlen += ruserlen; }
}
{	// write the remote host name
	unsigned short rhostlen = rhost?strlen(rhost):0;
	if (idtoklen<(cidlen+sizeof(unsigned long)+rhostlen)) {
		// not enough space in the allocated id token buffer--need to
		// reallocate
		if (rc = _increase_tknsize(cidlen+sizeof(unsigned long)+rhostlen, (void **)&idtokbuf, &idtoklen)) break;
	}
	// write the remote user id length... 
	*((unsigned short *)(idtokbuf+cidlen)) = htons(rhostlen); cidlen += sizeof(unsigned short);
	// ...and the remote user id
	if (rhostlen) { sprintf(idtokbuf+cidlen, "%s", rhost); cidlen += rhostlen; }
}
	// write the user id (len and value)
	if (!usrid) {
		/* build a user id, but for the time being return an error */
		// printf("Error [%s:%d]: no user id\n", __FILE__, __LINE__);
		rc = OSSH_NOUSERID_ERR; break;
	}
{	size_t usridlen = strlen(usrid);
	if (SHRT_MAX < usridlen) {
		// printf("Error [%s:%d]: userid too long\n", __FILE__, __LINE__);
		rc = OSSH_USERID_LEN_ERR; break;
	}
	if (idtoklen<(cidlen+sizeof(unsigned short)+usridlen)) {
		// not enough space in the allocated id token buffer--need to
		// reallocate
		if (rc = _increase_tknsize(cidlen+sizeof(unsigned short)+usridlen, (void **)&idtokbuf, &idtoklen)) break;
	}
	// write the user id length... 
	*((unsigned short *)(idtokbuf+cidlen)) =  htons((unsigned short)usridlen); cidlen += sizeof(unsigned short);
	// ...and the user id
	sprintf(idtokbuf+cidlen, "%s", usrid); cidlen += usridlen;
}
{	// generate an encrypted, 128bit AES session key; this will be done
	// only for RSA identity keys;
	// the whole idea with generating a session key is based on how
	// customers use such a large system like an HPC cluster; usually,
	// the user's home directory is mounted on the compute nodes, so
	// the application has access to the user's ~/.ssh subdir; when end
	// users setup key-based login, they add the OpenSSH public key to
	// their own authorized keys file and, on the compute node, they
	// use the public key from their own authorized keys file to login
	// as themselves on the compute node; however, in addition to the
	// authorized keys file, users also have access to their own
	// private key file, and that is what I am trying to exploit here;
	// of course, in the grand schema of things, there is still the
	// question of the network file system services accessing the end
	// user's home directory (and their private key, respectively)--if
	// that is not secure and the private key can be sniffed off the
	// network, what's the point having this key distibution mechanism?
	// one could ask;  the answer lays in what we try to protect here,
	// and that is separating users from each others and ensureing that
	// one end-user does not submit jobs as another end-user;
	// in addition to user's home directory setup, this mechanism
	// requires RSA keys;  to keep the key distribution simple and
	// scalable, the authn/authz mechanism employs a 1-to-many
	// communication path, i.e. one client to many servers;
	// consequentely, there is no negotiation between the client and
	// the server;  the server uses whatever the client sends, in one
	// shot, and that's it; all servers receive the same identity token
	// from the client; so the key is generated on the client side and
	// sent to the servers encrypted with the user's public key; the
	// servers decrypt the session key with the users's private key and
	// subsequently use it; the encryption and decryption with the
	// user's public and private key requires a key type that supports
	// such operations and the DSA standard does not support encryption
	// and decryption (that's not to say that one cannot use elliptic
	// keys to encrypt and decrypt data); consequently, the session key
	// can be distributed only if the user's identity file contains an
	// RSA key (i.e. it is either an SSH-1 or SSH-2 RSA identity file);
	size_t skeylen = 0; unsigned char *keybuf = NULL;
do {
	if (rsapub) {
		if (!(keybuf = malloc(2*(skeylen=OSSL_FNCSELECT(_RSA_size)(rsapub))))) {
			// unable to allocate memory for the encrypted key
			// printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
			rc = OSSH_MEMORY_ERR; break;
		}
		if (rc = _generate_session_key(rsapub, keybuf, &skeylen)) break;
	}
	// write the encrypted key and its length
	if (idtoklen<(cidlen+sizeof(unsigned short)+skeylen)) {
		// not enough space in the allocated id token buffer--need to
		// reallocate
		if (rc = _increase_tknsize(cidlen+sizeof(unsigned short)+skeylen, (void **)&idtokbuf, &idtoklen)) break;
	}
	*((unsigned short *)(idtokbuf+cidlen)) = htons((unsigned short)skeylen); cidlen += sizeof(unsigned short);
	if (skeylen) {
		memcpy((void *)(idtokbuf+cidlen), (void *)keybuf, skeylen);
		cidlen += skeylen;
	}
} while (0);
	if (keybuf) free(keybuf);
	if (rc) break;
}
{	// time stamp the id token
	Uu_int64_t tmstamp;
	tmstamp.l = time(NULL);
#if BYTE_ORDER == LITTLE_ENDIAN
	*((u_int32_t *)(idtokbuf+IDTOK_OFFST_TMSTAMP)) = htonl(tmstamp.i[1]);
	*((u_int32_t *)(idtokbuf+IDTOK_OFFST_TMSTAMP+sizeof(tmstamp.i[1]))) = htonl(tmstamp.i[0]);
#elif BYTE_ORDER == BIG_ENDIAN
	*((u_int32_t *)(idtokbuf+IDTOK_OFFST_TMSTAMP)) = htonl(tmstamp.i[0]);
	*((u_int32_t *)(idtokbuf+IDTOK_OFFST_TMSTAMP+sizeof(tmstamp.i[0]))) = htonl(tmstamp.i[1]);
#else
#error BYTE_ORDER must be either BIG_ENDIAN or LITTLE_ENDIAN
#endif
}
{	// sign the id token
	char idtoksha1[20] = "";
	unsigned int idtoksiglen = dsaprv?OSSL_FNCSELECT(_DSA_size)(dsaprv):OSSL_FNCSELECT(_RSA_size)(rsaprv);
	if (0==idtoksiglen) {
		// printf("Error [%s:%d]: invalid public key size...\n", __FILE__, __LINE__);
		rc = OSSH_DRSA_SIZE_ERR; break;
	}
	if (!OSSL_FNCSELECT(_SHA1)(idtokbuf, cidlen, idtoksha1)) {
		// printf("Error [%s:%d]: SHA1() failed for id token\n", __FILE__, __LINE__);
		rc = OSSH_SHA1_ERR; break;
	}
	if (idtoklen<cidlen+sizeof(unsigned short)+idtoksiglen) {
		// not enough space in the allocated id token buffer--need to
		// reallocate
		if (rc = _increase_tknsize(cidlen+sizeof(unsigned short)+idtoksiglen, (void **)&idtokbuf, &idtoklen)) break;
	}
	cidlen += sizeof(unsigned short);
	// write the signature...
	if (!(dsaprv?OSSL_FNCSELECT(_DSA_sign)(NID_sha1, idtoksha1, 20, (char *)(idtokbuf+cidlen), &idtoksiglen, dsaprv): OSSL_FNCSELECT(_RSA_sign)(NID_sha1, idtoksha1, 20, (char *)(idtokbuf+cidlen), &idtoksiglen, rsaprv))) {
		// printf("Error [%s:%d]: DSA/RSA_sign() failed...\n", __FILE__, __LINE__);
		rc = OSSH_DRSA_SIGN_ERR; break;
	}
	// ...and the signature length
	*((unsigned short *)(idtokbuf+cidlen-sizeof(unsigned short))) = htons((unsigned short)idtoksiglen); cidlen += idtoksiglen;
	// scale back the memory allocated for the id token buf. if need be
	if (cidlen<idtoklen) idtokbuf = realloc(idtokbuf, cidlen);
	idtok->iov_base = idtokbuf; idtok->iov_len = cidlen;
}
} while (0);
	if (rc) { memset(idtokbuf, 0, idtoklen); free(idtokbuf); }
}
} while (0);
	pthread_cleanup_pop(1);		// _nfree(usrid)
}
} while (0);
	// cleanup the public keys
	if (rsaprv) OSSL_FNCSELECT(_RSA_free)(rsaprv); if (rsapub) OSSL_FNCSELECT(_RSA_free)(rsapub);
	if (dsaprv) OSSL_FNCSELECT(_DSA_free)(dsaprv); if (dsapub) OSSL_FNCSELECT(_DSA_free)(dsapub);
}
} while (0);
	pthread_cleanup_pop(1);		// free(pubidfpath);
} while (0);
	pthread_cleanup_pop(1);		// free(idfpath);
}
} while (0);
	return rc;
}

#define IDTOK_STAT_VRFIED 0x80
#define IDTOK_STAT_ISVRFIED(idtok) \
	(IDTOK_STAT_VRFIED&(*((unsigned char *)(idtok->iov_base+IDTOK_OFFST_STATVER))))
#define IDTOK_STAT_SETVRFIED(idtok) \
	(*((unsigned char *)(idtok->iov_base+IDTOK_OFFST_STATVER)) |= IDTOK_STAT_VRFIED)
#define IDTOK_STAT_CLRVRFIED(idtok) \
	(*((unsigned char *)(idtok->iov_base+IDTOK_OFFST_STATVER)) &= ~IDTOK_STAT_VRFIED)
#define IDTOK_STAT_SKEY 0x40
#define IDTOK_STAT_HASKEY(idtok) \
	(IDTOK_STAT_SKEY&(*((unsigned char *)(idtok->iov_base+IDTOK_OFFST_STATVER))))
#define IDTOK_STAT_SETSKEY(idtok) \
	(*((unsigned char *)(idtok->iov_base+IDTOK_OFFST_STATVER)) |= IDTOK_STAT_SKEY)
#define IDTOK_STAT_CLRSKEY(idtok) \
	(*((unsigned char *)(idtok->iov_base+IDTOK_OFFST_STATVER)) &= ~IDTOK_STAT_SKEY)
int
psm__verify_id_token(
	char *uname,
	psm_idbuf_t idtok)
{
	int rc = 0;
do {
	if (!idtok || !idtok->iov_base || (IDTOK_LEN_MIN>=idtok->iov_len)) {
		// printf("Error [%s:%d]: invalid id buffer descriptor\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	}
	if (!uname||('\0'==uname[0])) {
		// printf("Error [%s:%d]: invalid user name\n", __FILE__, __LINE__);
		rc = OSSH_UNAME_INVALID_ERR; break;
	}
{	// check the id token
	size_t idtokbuflen = 0; unsigned char keytype = 0;
	char *cp = idtok->iov_base; int idtokver = 0;
	if (strncmp(cp, "ossh", strlen("ossh"))) {
		// printf("Error [%s:%d]: invalid id token\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	} else cp += strlen("ossh");
	if ((1!=(idtokver=0x0f&(*((unsigned char *)cp))))&&(2!=idtokver)) {
		// printf("Error [%s:%d]: invalid id token\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	} else cp += sizeof(unsigned char);
	if ((0x01!=(keytype = *cp))&&(0x02!=keytype)) {
		// printf("Error [%s:%d]: invalid id token\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	} else cp += sizeof(unsigned char);
	// skip the nonce
	cp += sizeof(Uu_int64_t);
{	// check the time against the skew
	time_t ltime = time(NULL);
	Uu_int64_t tmstamp = {0};
#if BYTE_ORDER == LITTLE_ENDIAN
	tmstamp.i[1] = ntohl(*((u_int32_t *)cp)); cp += sizeof(tmstamp.i[1]);
	tmstamp.i[0] = ntohl(*((u_int32_t *)cp)); cp += sizeof(tmstamp.i[0]);
#elif BYTE_ORDER == BIG_ENDIAN
	tmstamp.i[0] = ntohl(*((u_int32_t *)cp)); cp += sizeof(tmstamp.i[0]);
	tmstamp.i[1] = ntohl(*((u_int32_t *)cp)); cp += sizeof(tmstamp.i[1]);
#endif
	if (_idtokTTL&&(ltime>(time_t)tmstamp.l+_idtokTTL)) {
		// printf("Error [%s:%d]: id token skew too big\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_SKEW_ERR; break;
	}
}
{	// check the remote user id
	unsigned short tuserlen = ntohs(*((unsigned short *)cp)); cp += sizeof(unsigned short);
	if (tuserlen&&((tuserlen!=strlen(uname))||strncmp(uname, cp, tuserlen))) {
		// printf("Error [%s:%d]: id token not valid for target user %s\n", __FILE__, __LINE__, uname);
		rc = OSSH_IDTOK_USER_ERR; break;
	}
	cp += tuserlen;
}
{	// skip the host -- we'll come back to it, if need be
	unsigned short thostlen = ntohs(*((unsigned short *)cp)); cp += sizeof(unsigned short);
	cp += thostlen;
}
{
	char *rusrnid, idtoksha1[20] = "";
	unsigned short rusrnidlen = ntohs(*((unsigned short *)cp)); cp += sizeof(unsigned short);
	if (!rusrnidlen) {
		// printf("Error [%s:%d]: invalid id token\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	}
	if (idtok->iov_len<((void *)cp-idtok->iov_base+rusrnidlen)) {
		// printf("Error [%s:%d]: invalid id token\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	}
	if (!(rusrnid = malloc(rusrnidlen+1))) {
		// printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		rc = OSSH_MEMORY_ERR; break;
	}
	pthread_cleanup_push((void(*)(void *))free, rusrnid);
do {
	strncpy(rusrnid, cp, rusrnidlen); rusrnid[rusrnidlen] = '\0';
	// skip the target user id
	cp += rusrnidlen;
	// skip the session key, if any
	if (1<idtokver) {
		// id tokens version 2 and above have a session key field
		unsigned short skeylen = ntohs(*((unsigned short *)cp));
		cp += sizeof(unsigned short)+skeylen;
	}
	idtokbuflen = (size_t)((void *)cp-idtok->iov_base);
	// calculate the SHA1 digest
	if (!(OSSL_FNCSELECT(_SHA1)(idtok->iov_base, idtokbuflen, idtoksha1))) {
		// printf("Error [%s:%d]: SHA1() failed for id token\n", __FILE__, __LINE__);
		rc = OSSH_SHA1_ERR; break;
	}
{	// get the signature and its length
	unsigned short idtoksiglen = ntohs(*((unsigned short *)cp));
	unsigned char *idtoksig = (unsigned char *)(cp+=sizeof(unsigned short));
	if (!idtoksiglen) {
		// printf("Error [%s:%d]: invalid id token\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	}
	if (idtok->iov_len<((void *)cp-idtok->iov_base+idtoksiglen)) {
		// printf("Error [%s:%d]: invalid id token\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	}
{	// get the name of the authorized key file
	char *azfpath = NULL;
	if (rc = _get_authz_fname(uname, &azfpath)) {
		// printf("Error [%s:%d]: invalid authorized key file\n", __FILE__, __LINE__);
		break;
	}
	pthread_cleanup_push((void(*)(void *))free, azfpath);
do {
	// open the authorized keys file and read it, one line at a time
	FILE *azfstrm = NULL;
	if (!(azfstrm=fopen(azfpath, "r"))) {
		// printf("Error [%s:%d]: fopen() failed for %s: errno = %d\n", __FILE__, __LINE__, azfpath, errno);
		rc = OSSH_AUTHZFILE_OPEN_ERR; break;
	}
	pthread_cleanup_push((void(*)(void *))fclose, azfstrm);
do {
	char *cp, pubkeydata[8*1024] = "";
	do {
		if (!(cp = fgets(pubkeydata, 8*1024, azfstrm))) {
			// printf("Error [%s:%d]: fgets() failed or EOF\n", __FILE__, __LINE__);
			rc = OSSH_AUTHZFILE_GETS_ERR; break;
		}
		if ('#' != *cp) {
			DSA *dsapub = NULL; RSA *rsapub = NULL; char *usrid = NULL;
			// try to read the key data as RSA
			if (0x01==keytype) {
				if (!(dsapub = _read_dsa_public_key(cp, &usrid))) continue;
				if (usrid&&strcmp(usrid, rusrnid)) {
					// user ids don't match
					OSSL_FNCSELECT(_DSA_free)(dsapub); free(usrid); continue;
				}
				rc = OSSL_FNCSELECT(_DSA_verify)(NID_sha1, idtoksha1, 20, idtoksig, idtoksiglen, dsapub); OSSL_FNCSELECT(_DSA_free)(dsapub);
				if (!rc||(0>rc)) {
					rc = OSSH_DSA_VERIFY_ERR;
					// printf("Error [%s:%d]: DSA_verify() failed...\n", __FILE__, __LINE__);
					// there is no user id associated w/this key, so
					// I'll continue searching for another DSA key
					continue;
				} else rc = 0;
				if (usrid) free(usrid);
				break;
			}
			if (0x02==keytype) {
				if (!(rsapub = _read_rsa_public_key(cp, &usrid))) continue;
				if (usrid&&strcmp(usrid, rusrnid)) {
					// user ids don't match
					OSSL_FNCSELECT(_RSA_free)(rsapub); free(usrid); continue;
				}
				rc = OSSL_FNCSELECT(_RSA_verify)(NID_sha1, idtoksha1, 20, idtoksig, idtoksiglen, rsapub); OSSL_FNCSELECT(_RSA_free)(rsapub);
				if (!rc||(0>rc)) {
					rc = OSSH_RSA_VERIFY_ERR;
					// printf("Error [%s:%d]: RSA_verify() failed...\n", __FILE__, __LINE__);
					// there is no user id associated w/this key, so
					// I'll continue searching for another RSA key
					continue;
				} else rc = 0;
				if (usrid) free(usrid);
				break;
			}
		}
	} while (1);
	if (rc) {
		// just to make sure the verified bit is not on
		IDTOK_STAT_CLRVRFIED(idtok);
		break;
	} else IDTOK_STAT_SETVRFIED(idtok);
} while (0);
	pthread_cleanup_pop(1);		// fclose(azfstrm);
} while (0);
	pthread_cleanup_pop(1);		// free(azfpath);
}
}
} while (0);
	pthread_cleanup_pop(1);		// free(rusrnid);
}
}
} while (0);
	return rc;
}

int
psm__free_id_token(
	psm_idbuf_t idtok)
{
	int rc = 0;
do {
	if (!idtok || !idtok->iov_base || (0==idtok->iov_len)) {
		// printf("Error [%s:%d]: invalid id buffer descriptor\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	}
	free(idtok->iov_base);
	idtok->iov_base = (void *)NULL; idtok->iov_len = 0;
} while (0);
	return rc;
}

int
psm__get_id_from_token(
	psm_idbuf_t idtok,
	char *usrnid,
	size_t *usrnidlen)
{
	int rc = 0;
do {
	if (!idtok || !idtok->iov_base || (0==idtok->iov_len)) {
		// printf("Error [%s:%d]: invalid id buffer descriptor\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	}
	if (!usrnidlen) {
		// printf("Error [%s:%d]: invalid userid buffer length\n", __FILE__, __LINE__);
		rc = OSSH_ARG_ERR; break;
	}
	if (!IDTOK_STAT_ISVRFIED(idtok)) {
		// printf("Error [%s:%d]: identity token is not verified\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_NOTVERIFD_ERR; break;
	}
{	size_t tusrnidlen; char *tusrnid;
	size_t nidoffset = IDTOK_OFFST_IDDATA;
	// skip the target user name
	nidoffset += ntohs(*((unsigned short *)(idtok->iov_base+nidoffset)))+sizeof(unsigned short);
	// skip the target host name
	nidoffset += ntohs(*((unsigned short *)(idtok->iov_base+nidoffset)))+sizeof(unsigned short);
	// get the user name (i.e. the client's name)
	tusrnidlen = ntohs(*((unsigned short *)(idtok->iov_base+nidoffset)))+1;
	if ((*usrnidlen<tusrnidlen)||(!usrnid)) { *usrnidlen = tusrnidlen; rc = PSM__MEMORY_ERR; break; }
	tusrnid = (char *)(idtok->iov_base+nidoffset+sizeof(unsigned short)); \
	strncpy(usrnid, tusrnid, tusrnidlen-1); usrnid[tusrnidlen-1] = '\0';
	*usrnidlen = tusrnidlen;
}
} while (0);
	return rc;
}

int
psm__get_key_from_token(
	char *uname,
	psm_idbuf_t idtok,
	unsigned char *keybuf,
	size_t *keylen)
{
	int rc = 0;
do {
	if (!idtok || (NULL==idtok->iov_base) || (IDTOK_LEN_MIN>idtok->iov_len)) {
		// printf("Error [%s:%d]: invalid id buffer descriptor\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	}
	if (!keylen) {
		// printf("Error [%s:%d]: invalid key buffer length\n", __FILE__, __LINE__);
		rc = OSSH_ARG_ERR; break;
	}
	if (!keybuf || (16>*keylen)) {
		// printf("Error [%s:%d]: not enough memory for the session key\n", __FILE__, __LINE__);
		*keylen = 16; rc = PSM__MEMORY_ERR; break;
	}
	// if user argument is provided, then I expect the token to have
	// been verified first
	if (uname && ('\0'!=*uname) && !IDTOK_STAT_ISVRFIED(idtok)) {
		// printf("Error [%s:%d]: identity token is not verified\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_NOTVERIFD_ERR; break;
	}
	// check the identity token
	if (strncmp((char *)idtok->iov_base, "ossh", strlen("ossh"))) {
		// printf("Error [%s:%d]: invalid id token\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	}
{	unsigned char idtokver = 0x0f&(*((unsigned char *)(idtok->iov_base+IDTOK_OFFST_STATVER)));
	if (2!=idtokver) {
		// printf("Error [%s:%d]: invalid id token\n", __FILE__, __LINE__);
		rc = OSSH_IDTOK_INVALID_ERR; break;
	}
}
{	unsigned char keytype = *((unsigned char *)(idtok->iov_base+IDTOK_OFFST_KEYTYPE));
	if (0x02!=keytype) {
		// printf("Error [%s:%d]: invalid id token\n", __FILE__, __LINE__);
		rc = OSSH_SKEY_NOTRSA_ERR; break;
	}
}
{	// check the user name, if provided
	char *cp = idtok->iov_base+IDTOK_OFFST_IDDATA;
{	size_t tuserlen = ntohs(*((unsigned short *)cp)); cp += sizeof(unsigned short);
	if (uname && ('\0' != *uname)) {
		// make sure it's the same as the remote user in the id token
		if (tuserlen&&((tuserlen!=strlen(uname))||strncmp(uname, cp-sizeof(unsigned short), tuserlen))) {
			// printf("Error [%s:%d]: id token not valid for target user %s\n", __FILE__, __LINE__, uname);
			rc = OSSH_IDTOK_USER_ERR; break;
		}
	}
	cp += tuserlen;
}
	// skip the host name
	cp += ntohs(*((unsigned short *)cp)) + sizeof(unsigned short);
	// skip the remote user network id
	cp += ntohs(*((unsigned short *)cp)) + sizeof(unsigned short);
{	// check whether there is a session key
	// the public key used is RSA and there is a session key
	// get the user's identity file and try to de decrypt the
	// session key; if the user's identity file matches the public
	// keys used to initially encrypt the session key, then we can
	// decrypt the session key and will replace it w/the decrypted
	// value;  this should work as the session key should be a
	// 128bit AES key, which is smaller than its encrypted version
	// (which is the size of the RSA key, at least 512bits);
	// any error encountered in this code path will not result
	// in invalidating the identity token, because the token has
	// already been verified; however, it will rezult in no session
	// key and no subsequent message authentication services
	size_t ekeylen = ntohs(*((unsigned short *)cp));
	if (0 == ekeylen) {
		// printf("Error [%s:%d]: no session key in id token\n", __FILE__, __LINE__);
		rc = OSSH_SKEY_NOKEY_ERR; break;
	}
	cp += sizeof(unsigned short);
{	char *idfpath = NULL; RSA *rsaprv = NULL;
	if (rc = _get_identity_fname(uname, NULL, NULL, &idfpath)) break;
	pthread_cleanup_push((void(*)(void *))free, idfpath);
do {	// open the private identity file
	FILE *idfstrm = fopen(idfpath, "r");
	if (!idfstrm) {
		// printf("Error [%s:%d]: fopen() failed for %s: errno = %d\n", __FILE__, __LINE__, idfpath, errno);
		rc = OSSH_IDFILE_OPEN_ERR; break;
	}
	pthread_cleanup_push((void(*)(void *))fclose, idfstrm);
do {	// read the private key
	if (rsaprv = _read_rsa1_private_key(idfstrm)) break;
	// rewind the stream to the beginning
	fseek(idfstrm, 0, SEEK_SET);
	if (rsaprv = OSSL_FNCSELECT(_PEM_read_RSAPrivateKey)(idfstrm, NULL, NULL, "")) break;
	// printf("Error [%s:%d]: unable to read user's private key\n", __FILE__, __LINE__);
	rc = OSSH_IDFILE_READ_ERR; break;
} while (0);
	pthread_cleanup_pop(1);		// fclose(idfstrm);
} while (0);
	pthread_cleanup_pop(1);		// free(idfpath);
	if (rc) break;
	pthread_cleanup_push((void(*)(void *))OSSL_FNCSELECT(_RSA_free), rsaprv);
do {	// now that I have the private key, decrypt the session key
	if (rc = _recover_session_key(rsaprv, cp, ekeylen, keybuf, keylen)) break;
} while (0);
	pthread_cleanup_pop(1);		// OSSL_FNCSELECT(_RSA_free)
}
}
}
} while (0);
	return rc;
}

int
_generate_md5_digest(
	struct iovec *in,
	int cnt,
	unsigned char *digest)
{
	// assumes valid arguments
	int rc = 0;
	struct iovec *tdata = in;
do {
	MD5_CTX md5ctx = {0};
	// initialize the MD5 engine
	if (1 != OSSL_FNCSELECT(_MD5_Init)(&md5ctx)) {
		// printf("Error [%s:%d]: MD5_Init() failed...\n", __FILE__, __LINE__);
		rc = OSSH_MD5_INIT_ERR; break;
	}
	while (cnt--) {
		if (1 != OSSL_FNCSELECT(_MD5_Update)(&md5ctx, tdata->iov_base, tdata->iov_len)) {
			// printf("Error [%s:%d]: MD5_Update() failed...\n", __FILE__, __LINE__);
			rc = OSSH_MD5_UPDATE_ERR; break;
		}
		tdata++;
	}
	if (rc) break;
	if (1 != OSSL_FNCSELECT(_MD5_Final)(digest, &md5ctx)) {
		// printf("Error [%s:%d]: MD5_Final() failed...\n", __FILE__, __LINE__);
		rc = OSSH_MD5_FINAL_ERR; break;
	}
} while (0);
	return rc;
}

int
psm__sign_data(
	unsigned char *key,
	size_t keylen,
	struct iovec *in,
	int cnt,
	struct iovec *sig)
{
	// generate an MD5 digest of the messages received and then encrypt
	// the digest w/the 128bit AES key
	// data points to a NULL-terminated array of struct iovec scalars
	int rc = 0;
do {
	if (!in || (0 == cnt)) {
		// printf("Error [%s:%d]: invalid data vector\n", __FILE__, __LINE__);
		rc = OSSH_DATA_INVALID_ERR; break;
	}
{
	// generate the md5 digest
	unsigned char md5digest[MD5_DIGEST_LENGTH];
	if (rc = _generate_md5_digest(in, cnt, md5digest)) break;
{	// sign the digest w/the AES key
	AES_KEY keysched = {0};
	unsigned char *aescipher = malloc(AES_BLOCK_SIZE);
	if (!aescipher) {
		// printf("Error [%s:%d]: malloc() failed...\n", __FILE__, __LINE__);
		rc = OSSH_MEMORY_ERR; break;
	}
	if (rc = OSSL_FNCSELECT(_AES_set_encrypt_key)(key, 8*keylen, &keysched)) {
		// printf("Error [%s:%d]: AES_set_encrypt_key() failed...\n", __FILE__, __LINE__);
		rc = OSSH_AES_KEYSCHED_ERR; break;
	}
	OSSL_FNCSELECT(_AES_encrypt)(md5digest, aescipher, &keysched);
	sig->iov_base = aescipher; sig->iov_len = AES_BLOCK_SIZE;
}
}
} while (0);
	return rc;
}

int
psm__verify_data(
	unsigned char *key,
	size_t keylen,
	struct iovec *in,
	int cnt,
	struct iovec *sig)
{
	int rc = 0;
do {
	if (!sig||(!sig->iov_base)||(16!=sig->iov_len)) {
		// printf("Error [%s:%d]: signature is not valid\n", __FILE__, __LINE__);
		rc = OSSH_SIG_INVALID_ERR; break;
	}
{	struct iovec tsig = {0};
	if (rc = psm__sign_data(key, keylen, in, cnt, &tsig)) break;
	if ((tsig.iov_len!=sig->iov_len)||memcmp(tsig.iov_base, sig->iov_base, tsig.iov_len)) {
		// printf("Error [%s:%d]: signature verification failed\n", __FILE__, __LINE__);
		rc = OSSH_SIG_SIGNATURE_ERR;
	}
	psm__free_signature(&tsig);
}
} while (0);
	return rc;
}

int
psm__free_signature(
	struct iovec *sig)
{
	int rc = 0;
do {
	if (!sig || (0==sig->iov_len) || (!sig->iov_base)) {
		// printf("Error [%s:%d]: invalid signature descriptor\n", __FILE__, __LINE__);
		rc = OSSH_SIG_INVALID_ERR; break;
	}
	free(sig->iov_base);
	sig->iov_base = (void *)NULL; sig->iov_len = 0;
} while (0);
	return rc;
}

// EOF
