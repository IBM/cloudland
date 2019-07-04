#include<stdio.h>
#include<ctype.h>
#include<stdlib.h>
#include<string.h>
#include<unistd.h>
#include<signal.h>
#include<sys/wait.h>
#include<sys/types.h>
#include<sys/socket.h>
#include<netinet/in.h>
#include<arpa/inet.h>
#include<errno.h>
#include<assert.h>
#include<sys/file.h>
#include<sys/msg.h>
#include<sys/ipc.h>
#include<time.h>

struct dhcp_message
{
	uint8_t op;
	uint8_t htype;
	uint8_t hlen;
	uint8_t hops;
	uint32_t xid;
	uint16_t secs;
	uint16_t flags;
	uint32_t ciaddr;
	uint32_t yiaddr;
	uint32_t siaddr;
	uint32_t giaddr;
	char chaddr[16];
	char sname[64];
	char file[128];
	char magic[4];
	char opt[3];
} __attribute__((__packed__));

int send_dhcp_discovery(char *mac) 
{
	int rc, sockfd, yes, i, val;
	char *p, *sval, *tok, *stok;
	char savemac[16] = {0};
	struct sockaddr_in servaddr = {0};
	struct dhcp_message dhcpmsg = {0};

	assert(mac != NULL);
	strncpy(savemac, mac, sizeof(savemac));
	sockfd = socket(AF_INET, SOCK_DGRAM, 0);
	yes = 1;
	rc = setsockopt(sockfd, SOL_SOCKET, SO_BROADCAST, &yes, sizeof(yes));
	servaddr.sin_port = htons(67);
	servaddr.sin_family = AF_INET;
	servaddr.sin_addr.s_addr = inet_addr("255.255.255.255");

	dhcpmsg.op = 1;
	dhcpmsg.htype = 1;
	dhcpmsg.hlen = 6;
	dhcpmsg.hops = 0;
	val = random();
	dhcpmsg.xid = htonl(val);
	dhcpmsg.secs = htons(0);
	dhcpmsg.flags = 0; 
	p = mac;
	for (i = 0; i < 6; i++) {
		tok = strtok_r(p, ":", &stok);
		val = strtol(tok, &sval, 16);
		dhcpmsg.chaddr[i] = val;
		p = NULL;
	}
	dhcpmsg.magic[0]=99;
	dhcpmsg.magic[1]=130;
	dhcpmsg.magic[2]=83;
	dhcpmsg.magic[3]=99;
	dhcpmsg.opt[0]=53;
	dhcpmsg.opt[1]=1;
	dhcpmsg.opt[2]=1;
	rc = sendto(sockfd,&dhcpmsg,sizeof(dhcpmsg),0,(struct sockaddr*)&servaddr,sizeof(servaddr));

	return 0;
}

