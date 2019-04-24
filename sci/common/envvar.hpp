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
   10/06/08 nieyy        Initial code (D153875)

****************************************************************************/

#ifndef _ENVVAR_HPP
#define _ENVVAR_HPP

#include <map>
#include <string>

using namespace std;

class EnvVar 
{
    private:
        map<string, string> envlist;
        string              retval;
        
    public:
        EnvVar();
        ~EnvVar();
        
        void set(string &env, const char *val);
        void set(string &env, string &val);
        void set(string &env, int val);
        void set(string &env, long long val);
        void set(const char *env, const char *val);
        void set(const char *env, string &val);
        void set(const char *env, int val);
        void set(const char *env, long long val);
        string & get(string &env);
        string & get(const char *env);

        string & getExportcmd();
        string & getEnvString();

        void unsetAll();
        void dump();
};

#endif

