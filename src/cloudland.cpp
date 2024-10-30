/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

#include <sys/types.h>
#include <sys/stat.h>
#include <signal.h>
#include <sys/types.h>
#include <fcntl.h>
#include <string.h>
#include <unistd.h>
#include <stdlib.h>
#include <sci.h>

#include <fstream>

#include "handler.hpp"
#include "rpcworker.hpp"
#include "netlayer.hpp"
#include "log.hpp"

const int MAXFD = 1024;

string pidFile;
string hostList;
struct sigaction oldSa;

void sig_term(int sig)
{
    if (sig == SIGTERM) {
        log_crit("Terminating ... ");
        unlink(pidFile.c_str());
        exit(0);
    }
}

void set_oom_adj(int s)
{
    char oomfname[256] = "";
    char oomadjstr[16] = "";
    int oomfd = -1;
    struct stat oomadjst = {0};
    int nbytes = -1;

    int score = s;
    score = (score < -1000) ? (-1000) : (score);
    score = (score > 1000) ? (1000) : (score);

    ::sprintf(oomfname, "/proc/%d/oom_score_adj", getpid());
    if (::stat(oomfname, &oomadjst) != 0) {
        /* could not find oom_score_adj, will try oom_adj instead */
        ::sprintf(oomfname, "/proc/%d/oom_adj", getpid());
        double oomadjval = (score < 0) ? (((double) score/1000.0)*17.0) : (((double) score/1000.0)*15.0);
        ::sprintf(oomadjstr, "%.0f", oomadjval);
    } else {
        ::sprintf(oomadjstr, "%d", score);
    }

    oomfd = ::open(oomfname, O_WRONLY, 0);
    if (oomfd < 0) {
        log_error("open() failed for %s: errno = %d\n", oomfname, errno);
        return;
    }

    nbytes = ::write(oomfd, oomadjstr, strlen(oomadjstr));
    if (nbytes < 0) {
        log_error("write() failed for %s: errno = %d", oomfname, errno);
    } else {
        log_crit("wrote %d bytes to %s: %s", nbytes, oomfname, oomadjstr);
    }
    ::close(oomfd);
}

void daemonInit()
{
    struct sigaction sa;

    pid_t pid;
    if ((pid = fork()) < 0)
        exit(-1);
    else if (pid != 0) /* parent */
        exit(0);
    setsid();

    umask(0);
    ::sigaction(SIGCHLD, NULL, &oldSa);
    sa.sa_handler = SIG_IGN;
    sa.sa_flags = 0;
    sigemptyset(&sa.sa_mask);
    sigaction(SIGHUP, &sa, NULL);
    sigaction(SIGINT, &sa, NULL);
    sigaction(SIGPIPE, &sa, NULL);
    sigaction(SIGUSR1, &sa, NULL);
    sigaction(SIGUSR2, &sa, NULL);
    sigaction(SIGCHLD, &sa, NULL);
    sa.sa_handler = sig_term;
    sigaction(SIGTERM, &sa, NULL);

    if ((pid = fork()) < 0)
        exit(-1);
    else if (pid != 0) /* parent */
        exit(0);
    chdir("/");

    /* close off file descriptors */
    for (int i = 0; i < MAXFD; i++)
        close(i);

    /* redirect stdin, stdout, and stderr to /dev/null */
    open("/dev/null", O_RDONLY);
    open("/dev/null", O_RDWR);
    open("/dev/null", O_RDWR);
}

void writePidFile(string &pidf)
{
    unlink(pidf.c_str());
    ofstream pidfile(pidf.c_str());
    if (!pidfile) {
        printf("Cann't write pid file %s", pidf.c_str());
        return;
    }
    pidfile << (int)getpid();
}

int checkPidFile(string &pidf)
{
    ifstream pidfile(pidf.c_str());
    if (!pidfile)
        return 0;

    string line;
    pidfile >> line;
    if (line.size() == 0)
        return 0;

    int pid = atoi(line.c_str());
    if (kill(pid, 0) == 0)
        return -1;

    return 0;
}

void usage(char * pName)
{
    printf("%s [-p pidDir] [-l logDir] [-s severity] [-c hostFile] [-e] \n\n", pName);
    printf("    -p pidDir\t\tSpecify the pid directory. Default is: \"/var/run/\"(Linux) and \"/var/opt/\"(AIX)\n");
    printf("    -l logDir\t\tSpecify the log directory. Default is: \"/var/log/\"\n");
    printf("    -s severity\t\tSpecify the log severity. Default is: \"'3'(INFORMATION)\"\n");
    printf("    -e \t\t\tEnable the log. Default is: disabled\n");
    printf("    -c hostFile\t\t\tSpecify cloud host list file\n");
}

int initParams(int argc, char *argv[])
{
    int i;
    char *optpattern = "hl:p:s:c:";
    char *prog = argv[0];
    char *p = NULL;
    string logDir = "/opt/cloudland/log";
    int logLevel = Loger::INFORMATION;
    string pidDir = "/opt/cloudland/run";
    string logFile;

    extern char *optarg;
    p = strrchr(prog, '/');
    if (p != NULL)
        p++;
    else
        p = prog;

    hostList = "/etc/cloudland.hosts";
    while ((i = getopt(argc, argv, optpattern)) != EOF) {
        switch (i) {
            case 'l':
                logDir = optarg;
                break ;
            case 'p':
                pidDir = optarg;
                break;
            case 'c':
                hostList = optarg;
                break;
            case 's':
                logLevel = atoi(optarg);
                break;
            case 'h':
                usage(p);
                exit(0);
                break;
        }
    }

/*
    pidFile = pidDir + "/" + p + ".pid";
    if (checkPidFile(pidFile) < 0) {
        printf("%s is already running...\n", p);
        return -1;
    }
    daemonInit();
    writePidFile(pidFile);
*/

    logFile = string(p) + ".log";
    Loger::getInstance()->init(logDir.c_str(), logFile.c_str(), logLevel, Loger::ENABLE);
    set_oom_adj(-1000);

    return 0;
}

int main(int argc, char *argv[])
{
    if (initParams(argc, argv) != 0)
        return -1;

    RpcWorker worker;
    worker.runServer();

    return 0;
}
