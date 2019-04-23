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
#include <stdio.h>
#include <errno.h>
#include <unistd.h>
#include <string.h>
#include <assert.h>
#include <fcntl.h>
#include <pwd.h>
#include <sys/param.h>
#include <sys/stat.h>
#include <sys/uio.h>

#include "psec.h"

unsigned char testdata[] = {
0xde, 0xad, 0xbe, 0xef, 0x15, 0x04, 0x16, 0x89, 0x92, 0x73, 0x30, 0x68,
0x19, 0xa7, 0xb7, 0x2d, 0xde, 0xbe, 0xff, 0x00, 0x79, 0xc8, 0x8f, 0x5b,
0x30, 0x50, 0x67, 0x58, 0xdf, 0x11, 0x94, 0xe3, 0x8b, 0x39, 0x33, 0x94,
0x93, 0x83, 0xaa, 0x93, 0xa9, 0x93, 0xf2, 0x3d, 0x3e, 0x0d, 0x83, 0xf8,
0x38, 0x47, 0xd3, 0x8a, 0x85, 0x78, 0x82, 0x58, 0x56, 0xa9, 0x9a, 0x67,
0x01, 0x47, 0xab, 0xe7, 0x23, 0xc0, 0xf1, 0x38, 0xa7, 0xf0, 0x91, 0xc8};
struct iovec testvector[3] = {{(void *)testdata, 13}, {(void *)testdata+17, 23}, {(void *)testdata+50, 6}};

int
main(int argc, char *argv[])
{
	int rc = 0;
	unsigned int amdlhndl = 0;
	char amfpath[PATH_MAX] = "", idtokfpath[PATH_MAX] = "", *cp, *progname = ((cp=strrchr(argv[0], '/'))?cp++:argv[0]);
	int creatver = 0;
	struct iovec signature = {0};
	char *username = NULL;
do {
	if ((2!=argc)&&(4!=argc)&&(5!=argc)) {
		printf("Error [%s:%d]: invalid number of arguments\n", __FILE__, __LINE__);
		printf("Usage:\n	%s <auth module file>\n", progname);
		printf("or\n	%s <auth module file> [create|verify] <id token file>\n", progname);
		printf("or\n	%s {<auth module file>} verify <id token file> <user name>\n", progname);
		printf("\nwhere\n""	<auth module file> is relative to current directory\n""	<id token file> is relative to current directory\n""<user name> is the target user\n");
		rc = -1; break;
	}
	if (2<argc) {
do {
		if (('c'==*(argv[2]))&&(strstr("create", argv[2]))) {
			creatver = 1; break;
		}
		if (('v'==*(argv[2]))&&(strstr("verify", argv[2]))) {
			creatver = 2; break;
		}
		printf("Error [%s:%d]: invalid action: %s\n", __FILE__, __LINE__, argv[2]);
		rc = -1;
} while (0);
		if (rc) break;
	}
	if ((5==argc)&&(2==creatver)) username = argv[4];
	// get current directory
	if (0>getcwd(amfpath, PATH_MAX)) {
		printf("Error [%s:%d]: getcwd() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		rc = -1; break;
	}
	if ('/' != amfpath[strlen(amfpath)]) strcat(amfpath, "/");
	strcpy(idtokfpath, amfpath);
	strcat(amfpath, argv[1]);
	if (rc = psec_set_auth_module(NULL, amfpath, "m[t=-1]", &amdlhndl)) {
		printf("Error [%s:%d]: psec_set_auth_module() failed\n", __FILE__, __LINE__);
		rc = -1; break;
	}
{
	psec_idbuf_desc idbufd = {0};
	if (2!=creatver) {
		// we need to get an id token
		if (rc = psec_get_id_token(amdlhndl, NULL, NULL, &idbufd)) {
			printf("Error [%s:%d]: psec_get_id_token() failed with rc = %d\n", __FILE__, __LINE__, rc);
			rc = -1; break;
		}
		if (1==creatver) {
			// save the id token in the file provided and exit
			strcat(idtokfpath, argv[3]);
{
			int idtokfd = -1; size_t nbytes = -1;
			if (0>(idtokfd = open(idtokfpath, O_WRONLY|O_CREAT|O_EXCL, S_IRUSR|S_IWUSR))) {
				printf("Error [%s:%d]: open() failed w/errno = %d\n", __FILE__, __LINE__, errno);
				rc = -1; break;
			}
do {
			if (0>(nbytes=write(idtokfd, idbufd.iov_base, idbufd.iov_len))) {
				printf("Error [%s:%d]: write() failed w/errno = %d\n", __FILE__, __LINE__, errno);
				rc = -1; break;
			}
			if (nbytes!=idbufd.iov_len) {
				printf("Error [%s:%d]: failure to write entire id token to file\n", __FILE__, __LINE__);
				rc = -1; break;
			}
} while (0);
			close(idtokfd);
}
			break;
		} else {
			// get the session key
			unsigned char skey[10], *skeyp = skey; size_t skeylen = 10;
do {
do {
			if (!(rc = psec_get_key_from_token(amdlhndl, NULL, &idbufd, skeyp, &skeylen))) break; 
			if (PSEC_MEMORY_ERR!=rc) {
				printf("Error [%s:%d]: psec_get_key_from_token() failed: rc = %d\n", __FILE__, __LINE__);
				break;
			}
			if (!skeylen) {
				printf("Error [%s:%d]: psec_get_key_from_token() failed: rc = %d\n", __FILE__, __LINE__);
				break;
			}
			skeyp = malloc(skeylen); assert(skeyp);
			if (rc = psec_get_key_from_token(amdlhndl, NULL, &idbufd, skeyp, &skeylen)) {
				printf("Error [%s:%d]: psec_get_key_from_token() failed: rc = %d\n", __FILE__, __LINE__);
				break;
			}
} while (0);
			if (rc) break;
			// use some test data to sign
			if (rc = psec_sign_data(amdlhndl, skeyp, skeylen, testvector, 3, &signature)) {
				printf("Error [%s:%d]: psec_sign_data() failed: rc = %d\n", __FILE__, __LINE__);
				break;
			}
{			// one more time, for consistency
			struct iovec sig2 = {0};
			if (rc = psec_sign_data(amdlhndl, skeyp, skeylen, testvector, 3, &sig2)) {
				printf("Error [%s:%d]: psec_sign_data() failed: rc = %d\n", __FILE__, __LINE__);
				break;
			}
			if (memcmp(sig2.iov_base, signature.iov_base, sig2.iov_len)) {
				printf("Error [%s:%d]: psec_sign_data() failed with inconsistent results\n", __FILE__, __LINE__);
				rc = -1; break;
			}
}
} while (0);
			if (skeyp != skey) free (skeyp);
			if (rc) break;
		}
	}
do {
	// get the current user name
	long pwrbufsize = sysconf(_SC_GETPW_R_SIZE_MAX);
	void *pwrbuf = malloc(pwrbufsize); assert(pwrbuf);
do {
	struct passwd usrpwd, *usrpwdp = NULL;
	if (rc = getpwuid_r(geteuid(), &usrpwd, pwrbuf, pwrbufsize, &usrpwdp)) {
		printf("Error [%s:%d]: getpwuid_r() failed: rc = %d\n", __FILE__, __LINE__, rc);
 		rc = -1; break;
	}
	if (!usrpwd.pw_name) {
		printf("Error: [%s:%d]: no user name available\n", __FILE__, __LINE__);
		rc = -1; break;
	}
	if (2==creatver) {
		// read the id token from file
		strcat(idtokfpath, argv[3]);
{
		size_t nbytes; int idtokfd = -1; struct stat idtokfstat = {0};
		if (0>stat(idtokfpath, &idtokfstat)) {
			printf("Error [%s:%d]: stat() failed w/errno = %d for %s\n", __FILE__, __LINE__, errno, idtokfpath);
			rc = -1; break;
		}
		if (!idtokfstat.st_size) {
			printf("Error [%s:%d]: invalid id token file: %s\n", __FILE__, __LINE__, idtokfpath);
			rc = -1; break;
		}
		if (0>(idtokfd = open(idtokfpath, O_RDONLY))) {
			printf("Error [%s:%d]: open() failed w/errno = %d for %s\n", __FILE__, __LINE__, errno, idtokfpath);
			rc = -1; break;
		}
do {
		if (!(idbufd.iov_base = malloc(idbufd.iov_len = idtokfstat.st_size))) {
			printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
			rc = -1; break;
		}
		if (0>(nbytes = read(idtokfd, idbufd.iov_base, idbufd.iov_len))) {
			printf("Error [%s:%d]: read() failed w/errno = %d for %s\n", __FILE__, __LINE__, errno, idtokfpath);
			rc = -1; break;
		}
		if (nbytes!=idbufd.iov_len) {
			printf("Error [%s:%d]: failed to read entire id token from %s\n", __FILE__, __LINE__, idtokfpath);
			rc = -1; break;
		}
} while (0);
		close(idtokfd);
		if (rc) break;
}
	}
	if (rc = psec_verify_id_token(amdlhndl, username?username:usrpwd.pw_name, &idbufd)) {
		printf("Error [%s:%d]: psec_verify_id_token() failed\n", __FILE__, __LINE__);
		rc = -1; break;
	}
	printf("Info [%s:%d]: id token verified :-)\n", __FILE__, __LINE__);
{	// get the user's NID from the id token
	char usrnid[10] = "", *cp = usrnid; size_t usrnidlen = 10;
	int freeusrnid = 0;
do {
	if (!(rc = psec_get_id_from_token(amdlhndl, &idbufd, cp, &usrnidlen))) break;
	if ((PSEC_MEMORY_ERR!=rc)||(0==usrnidlen)) {
		// memory error; break out of here
		rc = -1; break;
	}
	// not enough memory provided for the user NID
	// usrnidlen should have the size of memory we need to allocate
	if (!(cp = malloc(usrnidlen))) {
		printf("Error [%s:%d]: malloc() failed w/errno = %d\n", __FILE__, __LINE__, errno);
		rc = -1; break;
	}
	freeusrnid++;
	if (rc = psec_get_id_from_token(amdlhndl, &idbufd, cp, &usrnidlen)) {
		printf("Error [%s:%d]: psec_get_id_from_token() failed w/rc = %d\n", __FILE__, __LINE__, rc);
		rc = -1; break;
	}
} while (0);
	if (!rc) printf("Info [%s:%d]: user NID = %s\n", __FILE__, __LINE__, cp);
	if (freeusrnid) free(cp);
}
{	// get the session key from the token and verify the signature
	unsigned char skey[20], *skeyp = skey; size_t skeylen = 20;
do {
	if (!(rc = psec_get_key_from_token(amdlhndl, username?username:usrpwd.pw_name, &idbufd, skeyp, &skeylen))) break;
	if (PSEC_MEMORY_ERR!=rc) {
		printf("Error [%s:%d]: psec_get_key_from_token() failed: rc = %d\n", __FILE__, __LINE__);
		rc = -1; break;
	}
	if (!skeylen) {
		printf("Error [%s:%d]: psec_get_key_from_token() failed: rc = %d\n", __FILE__, __LINE__);
		rc = -1; break;
	}
	skeyp = malloc(skeylen); assert(skeyp);
	if (rc = psec_get_key_from_token(amdlhndl, username?username:usrpwd.pw_name, &idbufd, skeyp, &skeylen)) {
		printf("Error [%s:%d]: psec_get_key_from_token() failed: rc = %d\n", __FILE__, __LINE__);
		rc = -1; break;
	}
} while (0);
	if (rc) break;
	if (2!=creatver) {
		// verify signature
		if (rc = psec_verify_data(amdlhndl, skeyp, skeylen, testvector, 3,  &signature)) {
			printf("Error [%s:%d]: psec_verify_data() failed w/rc = %d\n", __FILE__, __LINE__, rc);
			rc = -1; break;
		}
	}
}
} while (0);
	free(pwrbuf);
} while (0);
	if ((2==creatver)&&idbufd.iov_base) free(idbufd.iov_base);
	else psec_free_id_token(amdlhndl, &idbufd);
}
} while (0);
	if (rc = psec_free_signature(amdlhndl, &signature)) {
		printf("Error [%s:%d]: psec_free_signature() failed w/rc = %d\n", __FILE__, __LINE__, rc);
	}
	return rc;
}
