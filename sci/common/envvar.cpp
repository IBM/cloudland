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

 Classes: Envvar

 Description: Environment variable manipulation.
   
 Author: Nicole Nie, Liu Wei

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 nieyy        Initial code

****************************************************************************/

#include <assert.h>
#include <stdio.h>

#include "envvar.hpp"
#include "tools.hpp"

EnvVar::EnvVar()
{
}

EnvVar::~EnvVar()
{
    envlist.clear();
}

void EnvVar::set(string &env, const char *value)
{
    if (value) {
        envlist[env] = env + "=" + value;
    }
}

void EnvVar::set(string &env, string &value)
{
    envlist[env] = env + "=" + value;
}

void EnvVar::set(string &env, int value)
{
    envlist[env] = env + "=" + SysUtil::itoa(value);
}

void EnvVar::set(string &env, long long value)
{
    envlist[env] = env + "=" + SysUtil::lltoa(value);
}

void EnvVar::set(const char *env, const char *value)
{
    assert(env);
    if (value) {
        envlist[env] = string(env) + "=" + value;
    }
}

void EnvVar::set(const char *env, string &value)
{
    assert(env);
    envlist[env] = string(env) + "=" + value;
}

void EnvVar::set(const char *env, int value)
{
    assert(env);
    envlist[env] = string(env) + "=" + SysUtil::itoa(value);
}

void EnvVar::set(const char *env, long long value)
{
    assert(env);
    envlist[env] = string(env) + "=" + SysUtil::lltoa(value);
}

string & EnvVar::get(string &env)
{
    retval = "";
    
    if (envlist.find(env) != envlist.end())
        retval = envlist[env];
    
    return retval;
}

string & EnvVar::get(const char *env)
{
    retval = "";
    
    if (envlist.find(env) != envlist.end())
        retval = envlist[env];
    
    return retval;
}

string & EnvVar::getExportcmd()
{
    retval = "";
    
    map<string, string>::const_iterator p = envlist.begin();
    for(; p != envlist.end(); ++p) 
        retval += "export " + p->second + ";";

    return retval;
}

string & EnvVar::getEnvString()
{
    retval = "";
    
    map<string, string>::const_iterator p = envlist.begin();
    for(; p != envlist.end(); ++p) {
        retval += ";" + p->second;
    }
    
    return retval;
}

void EnvVar::unsetAll()
{
    envlist.clear();
}

void EnvVar::dump()
{
    map<string, string>::const_iterator p = envlist.begin();
    for(; p != envlist.end(); ++p) {
        printf("%s\n", p->second.c_str());
    }
}

