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
   
 Author: Liu Wei, Tu HongJ

 History:
   Date     Who ID    Description
   -------- --- ---   -----------
   10/06/08 lwbjcdl      Initial code (D153875)

****************************************************************************/

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif
#ifndef _LOG_HPP
#define _LOG_HPP

#include <pthread.h>
#include <string>

#define MAX_LOG_LEN 1024
#define MAX_PATH_LEN 512

using namespace std;

class Log 
{
    public:
        enum LEVEL {
            CRITICAL,
            ERROR,
            WARNING,
            INFORMATION,
            DEBUG,
            PERFORMANCE,
            OTHER
        };

        enum MODE {
            INVALID = -1,
            DISABLE = 0,
            ENABLE = 1
        };
        
    private:
        Log();
       
        int mode; 
        int permitLevel;
        string logDir;
        string logFile;
        char logPath[2 * MAX_PATH_LEN];

        static Log *logger;
        
    public:
        ~Log();
        static Log * getInstance() {
            if(logger == NULL)
                logger = new Log();
            return logger;
        }
        
        void init(const char *directory = "../log", const char * filename = "debug.log", int level = INFORMATION, int m = DISABLE);
        void rename(const char *directory = "../log", int level = INFORMATION, int m = INVALID);
        void print(int level, char * srcFile, int srcLine, const char * format, ...);

        string & getLogDir() { return logDir; }
        int getLogLevel() { return permitLevel; }
};

#ifdef _SCI_DEBUG

#define log_init(a, b, c, d)  Log::getInstance()->init(a, b, c, d)
#define log_rename(a, b, c)   Log::getInstance()->rename(a, b, c)
#define log_crit(...)      Log::getInstance()->print(Log::CRITICAL, __FILE__, __LINE__,  __VA_ARGS__)
#define log_error(...)     Log::getInstance()->print(Log::ERROR, __FILE__, __LINE__,  __VA_ARGS__)
#define log_warn(...)      Log::getInstance()->print(Log::WARNING, __FILE__, __LINE__,  __VA_ARGS__)
#define log_info(...)      Log::getInstance()->print(Log::INFORMATION, __FILE__, __LINE__,  __VA_ARGS__)
#define log_debug(...)     Log::getInstance()->print(Log::DEBUG, __FILE__, __LINE__,  __VA_ARGS__)
#define log_perf(...)      Log::getInstance()->print(Log::PERFORMANCE, __FILE__, __LINE__,  __VA_ARGS__)

#else

#define log_init(...)   
#define log_rename(...)   
#define log_crit(...)
#define log_error(...)
#define log_warn(...)
#define log_info(...)
#define log_debug(...)
#define log_perf(...)

#endif

#endif

