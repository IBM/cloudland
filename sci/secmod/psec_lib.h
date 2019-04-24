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

typedef struct __dlink_elem_s {
	struct __dlink_elem_s *prev;
	struct __dlink_elem_s *next;
} __dlink_elem_desc, *__dlink_elem_t;

void __rm_elem_from_dllist(__dlink_elem_t, __dlink_elem_t *);
void __add_elem_to_dllist(__dlink_elem_t, __dlink_elem_t *);
void __insert_elem_before_dllist(__dlink_elem_t, __dlink_elem_t *);

typedef struct _psec_module {
	struct _psec_module *prev;
	struct _psec_module *next;
	unsigned int psm_flags;
	unsigned int psm_stindex;
#define PSM_STATE_MASK 0x0000ffff
#define PSM_STATE_LOADED 0x0001
#define PSM_STATE_INITED 0x0002
	char *psm_opts;
	char *psm_name;
	char *psm_fpath;
	void *psm_object;
#define PSM_SYMTBLE_SIZE 10
	void *psm_symtble[PSM_SYMTBLE_SIZE];
	pthread_mutex_t psm_mutex;
} _psec_module_sec, *_psec_module_t;

typedef struct _psec_state {
	unsigned int pss_state;
	pthread_mutex_t pss_mutex;
	int pss_modcnt;
	_psec_module_t *pss_mdlslist;
	_psec_module_t pss_modules;
} _psec_state_desc, *_psec_state_t;

extern _psec_state_desc _PSEC_STATE;
