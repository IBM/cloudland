/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#ifndef _LOG_HPP
#define _LOG_HPP

#include <pthread.h>
#include <string>

#define MAX_LOG_LEN 1024
#define MAX_PATH_LEN 512

using namespace std;

class Loger 
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
        Loger();
       
        int mode; 
        int permitLevel;
        string logDir;
        string logFile;
        char logPath[2 * MAX_PATH_LEN];

        static Loger *logger;
        
    public:
        ~Loger();
        static Loger * getInstance() {
            if(logger == NULL)
                logger = new Loger();
            return logger;
        }
        
        void init(const char *directory = "../log", const char * filename = "debug.log", int level = INFORMATION, int m = DISABLE);
        void rename(const char *directory = "../log", int level = INFORMATION, int m = INVALID);
        void print(int level, char * srcFile, int srcLine, const char * format, ...);

        string & getLogDir() { return logDir; }
        int getLogLevel() { return permitLevel; }
};

#define log_init(a, b, c, d)  Loger::getInstance()->init(a, b, c, d)
#define log_rename(a, b, c)   Loger::getInstance()->rename(a, b, c)
#define log_crit(...)      Loger::getInstance()->print(Loger::CRITICAL, __FILE__, __LINE__,  __VA_ARGS__)
#define log_error(...)     Loger::getInstance()->print(Loger::ERROR, __FILE__, __LINE__,  __VA_ARGS__)
#define log_warn(...)      Loger::getInstance()->print(Loger::WARNING, __FILE__, __LINE__,  __VA_ARGS__)
#define log_info(...)      Loger::getInstance()->print(Loger::INFORMATION, __FILE__, __LINE__,  __VA_ARGS__)
#define log_debug(...)     Loger::getInstance()->print(Loger::DEBUG, __FILE__, __LINE__,  __VA_ARGS__)
#define log_perf(...)      Loger::getInstance()->print(Loger::PERFORMANCE, __FILE__, __LINE__,  __VA_ARGS__)

#endif

