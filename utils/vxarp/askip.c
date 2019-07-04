#include <sys/types.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <time.h>

#define SERVER_SOCKET "/var/run/get_ip_server.sock"
#define TRX_LEN 32

int main(int argc, char *argv[]) 
{ 
	int sock, rc; 
	struct sockaddr_un name;
	char buf[TRX_LEN] = {0};
	int slen = 0;
	time_t time1;
	struct tm tm1;
	char cli_sock[128] = {0};

	time(&time1);
	localtime_r(&time1, &tm1);
	strftime(cli_sock, sizeof(cli_sock), "/tmp/%y%m%d-%H:%M:%S.sock", &tm1);

	if (argc < 2) {
		fprintf(stderr, "%s <IP | MAC>\n", argv[0]);
		exit(1);
	}
	if (strlen(argv[1]) >= (TRX_LEN-1)) {
		fprintf(stderr, "Invalid argument!\n");
		exit(1);
	}

	sock = socket(AF_UNIX, SOCK_DGRAM, 0); 
	if (sock < 0) { 
		perror("Opening datagram socket"); 
		exit(1); 
	} 

	name.sun_family = AF_UNIX; 
	strcpy(name.sun_path, cli_sock); 
	rc = bind(sock, (void*)&name, sizeof(name));
	strcpy(name.sun_path, SERVER_SOCKET); 
	strcpy(buf, argv[1]);
	if (sendto(sock, buf, sizeof(buf), 0, (struct sockaddr *)&name, sizeof(struct sockaddr_un)) < 0) 
	{ 
		perror("Sending datagram message"); 
	} 
	memset(buf, '\0', sizeof(buf));
	slen = sizeof(name);
	recvfrom(sock, buf, sizeof(buf), 0, (struct sockaddr *)&name, (socklen_t *)&slen);
	printf("%s", buf);
	close(sock); 
	unlink(cli_sock);

	return 0;
} 
