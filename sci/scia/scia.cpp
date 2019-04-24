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

 Classes: None

 Description: main() function.
   
 Author: Nicole Nie, Tu HongJ, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 nieyy        Initial code (D153875)

****************************************************************************/

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <stdlib.h>
#include <stdio.h>
#include <stdarg.h>
#include <string.h>
#include <signal.h>
#include <unistd.h>
#include <ctype.h>
#include <assert.h>
#include <dlfcn.h>

#include "sci.h"

#ifdef _SCI_LINUX
#define SCI_LIB_PATH "libsci.so"
#else  // aix
#ifdef __64BIT__
#define SCI_LIB_PATH "libsci_r.a(libsci64_r.o)"
#else  // 32-bit
#define SCI_LIB_PATH "libsci_r.a(libsci_r.o)"
#endif // 32-bit
#endif // aix

typedef int (scia_init_hndlr)(sci_info_t *);
typedef int (scia_term_hndlr)();

int main()
{
    void *dlopen_file = NULL;
    char *error = NULL;
    int rc;

#if defined(_SCI_LINUX)
    dlopen_file = ::dlopen(SCI_LIB_PATH, RTLD_NOW | RTLD_LOCAL);
#elif defined(__APPLE__)
    dlopen_file = ::dlopen(SCI_LIB_PATH, RTLD_NOW | RTLD_LOCAL);
#else  // aix
    dlopen_file = ::dlopen(SCI_LIB_PATH, RTLD_NOW | RTLD_LOCAL | RTLD_MEMBER);
#endif
    if (!dlopen_file) {
        ::fprintf (stderr, "%s\n", ::dlerror());
        ::exit(1);
    }

    ::dlerror();    /* Clear any existing error */
    scia_init_hndlr *init_hndlr = (scia_init_hndlr *) ::dlsym(dlopen_file, "SCI_Initialize");
    scia_term_hndlr *term_hndlr = (scia_term_hndlr *) ::dlsym(dlopen_file, "SCI_Terminate");
    if ((error = ::dlerror()) != NULL)  {
        ::fprintf (stderr, "%s\n", error);
        ::exit(1);
    }
    
    // sleep(20);
    rc = init_hndlr(NULL);
    if (rc != SCI_SUCCESS) {
        ::fprintf(stderr, "scia initialization not perfect\n");
    }
    rc = term_hndlr();
    if (rc != SCI_SUCCESS) {
        ::fprintf(stderr, "scia termination failed\n");
        ::exit(1);
    }

    ::dlclose(dlopen_file);

    return 0;
}

