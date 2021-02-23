#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <ctype.h>
#include <fcntl.h>
#include <signal.h>
#include <sys/types.h>
#include <sys/stat.h>

#include "server.h"

void usage()
{
	printf("Usage: vxresolver [-p port] [-d] <-b dbfile>\n");
	printf("  -p port: listening port\n");
	printf("  -d : running as daemon\n");
	printf("  -b dbfile: sqlite3 db file\n");
}

void daemonize()
{
	int i = 0;
	pid_t   pid;
	struct sigaction sa;

	umask(0);

	if ((pid = fork()) < 0)
		exit(-1);
	else if (pid != 0) /* parent */
		exit(0);
	setsid();

	sa.sa_handler = SIG_IGN;
	sigemptyset(&sa.sa_mask);
	sa.sa_flags = 0;
	sigaction(SIGHUP, &sa, NULL);
	sigaction(SIGUSR1, &sa, NULL);
	sigaction(SIGUSR2, &sa, NULL);
	/*
	   sigaction(SIGTERM,&sa,NULL);
	   sigaction(SIGINT,&sa,NULL);
	   */
	if ((pid = fork()) < 0)
		exit(-1);
	else if (pid != 0) /* parent */
		exit(0);

	chdir("/");
	/* close off file descriptors */
	for (i = 0; i < MAXFD; i++)
		close(i);

	/* redirect stdin, stdout, and stderr to /dev/null */
	open("/dev/null", O_RDONLY);
	open("/dev/null", O_RDWR);
	open("/dev/null", O_RDWR);
}

int main(int argc, char *argv[])
{
	int i = -1;
	int rc;
	int vx_port = VXH_PORT;
	int dst_port = VXLAN_PORT;
	int db_type = SQLITE3;
	char *db_url = NULL;
	char *optpattern = "hp:d:t:x:";
	extern char *optarg;
	extern int  optind;

	while ((i = getopt(argc, argv, optpattern)) != EOF) {
		switch (i) {
			case 'p':
				vx_port = atoi(optarg);
				break;
			case 'x':
				dst_port = atoi(optarg);
				break;
			case 't':
				rc = strcmp(optarg, "postgres");
				if (rc == 0) {
					db_type = POSTGRES;
				}
				break;
			case 'd':
				db_url = optarg;
				break;
			case 'h':
				usage();
				return 1;
		}
	}
	if ((vx_port <= 0) || (dst_port <= 0)|| (db_url == NULL)) {
		usage();
		exit(1);
	}

	working(vx_port, dst_port, db_type, db_url);
	return 0;
}
