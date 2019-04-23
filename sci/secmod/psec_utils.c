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

#include "psec_lib.h"

/*- Double linked list routines -------------------------------------*/

void
__rm_elem_from_dllist(
	__dlink_elem_t elem,
	__dlink_elem_t *queue)
{
	/* assumes elem and queue are valid */
	if (elem->next) {
		if (elem->prev) {
			/* somewhere in the middle */
			elem->prev->next = elem->next;
			elem->next->prev = elem->prev;
		} else {
			/* the first one */
			*queue = elem->next;
			elem->next->prev = NULL;
		}
	} else {
		if (elem->prev) {
			/* the last one */
			elem->prev->next = NULL;
		} else {
			/* the only one */
			*queue = NULL;
		}
	}
	/* reset next and prev pointer */
	elem->prev = elem->next = NULL;
}

void
__add_elem_to_dllist(
	__dlink_elem_t elem,
	__dlink_elem_t *queue)
{
	/* assumes elem is valid */
	elem->next = *queue; elem->prev = NULL;
	if (*queue) (*queue)->prev = elem;
	*queue = elem;
}

void
__insert_elem_before_dllist(
	__dlink_elem_t elem,
	__dlink_elem_t *queue)
{
	/* assumes elem and queue are valid */
	__dlink_elem_t tmp = (*queue)->prev;
	if (tmp) {
		tmp->next = elem;
		elem->next = *queue; elem->prev = tmp;
		(*queue)->prev = elem;
	} else {
		elem->next = *queue; elem->prev = NULL;
		(*queue)->prev = elem; *queue = elem;
	}
}
