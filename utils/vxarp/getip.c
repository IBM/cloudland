#include <asm/types.h>
#include <sys/socket.h>
#include <sys/time.h>
#include <time.h>
#include <linux/netlink.h>
#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <netdb.h>
#include <string.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <signal.h>
#include <errno.h>
#include <assert.h>
#include <fcntl.h>
#include <netinet/if_ether.h>
#include <netinet/ether.h>
#include <netinet/ip.h>
#include <netinet/ip_icmp.h>
#include <linux/if.h>
#include <linux/netfilter_bridge/ebt_ulog.h>
#include <linux/netfilter_bridge.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>

#define VLAN_HLEN 4
#define BUFFER_LEN 65536
#define MAXFD 1024

extern int init_lookup_socket();
extern int insert(char *ip, char *mac, char *bridge);
extern int reply_query(int sock);
extern int check_entries();

struct vlan_hdr {
    unsigned short TCI;
    unsigned short encap;
};

struct dhcp_message {
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

static struct sockaddr_nl sa_local =
{
    .nl_family = AF_NETLINK,
    .nl_groups = 1,
};

static char buffer[BUFFER_LEN];
static int nl_sock;

ebt_ulog_packet_msg_t *get_ulog_packet()
{
    static struct nlmsghdr *nl_header = NULL;
    static int len, remain_len;
    ebt_ulog_packet_msg_t *msg;
    socklen_t addrlen = sizeof(sa_local);

    len = recvfrom(nl_sock, buffer, BUFFER_LEN, 0,
            (struct sockaddr *)&sa_local, &addrlen);
    if (len < 0) 
        return NULL;
    nl_header = (struct nlmsghdr *)buffer;
    if ((nl_header->nlmsg_flags & MSG_TRUNC) || (len > BUFFER_LEN)) 
        return NULL;
    if (!NLMSG_OK(nl_header, BUFFER_LEN)) 
        return NULL;

    msg = NLMSG_DATA(nl_header);

    remain_len = (len - ((char *)nl_header - buffer));
    if ((nl_header->nlmsg_flags & NLM_F_MULTI) && (nl_header->nlmsg_type != NLMSG_DONE))
        nl_header = NLMSG_NEXT(nl_header, remain_len);
    else
        nl_header = NULL;

    return msg;
}

void usage()
{
    printf("Usage: getip <group_number>\nwhere 0 < group_number < 33\n");
    exit(-1);
}

int init_ulog_socket(int group)
{
    int rc = -1;
    int rcvbufsize = BUFFER_LEN;
    int val = 0;

    sa_local.nl_groups = 1 << (group - 1);
    sa_local.nl_pid = getpid();
    nl_sock = socket(PF_NETLINK, SOCK_RAW, NETLINK_NFLOG);
    if (nl_sock < 0) {
        perror("socket");
        exit(-1);
    }
    rc = bind(nl_sock, (struct sockaddr *)(&sa_local), sizeof(sa_local));
    if (rc < 0) {
        perror("bind");
        exit(-1);
    }
    rc = setsockopt(nl_sock, SOL_SOCKET, SO_RCVBUF, &rcvbufsize, sizeof(rcvbufsize));
    val = fcntl(nl_sock, F_GETFL, 0);
    if (val != -1)
        rc = fcntl(nl_sock, F_SETFL, val | O_NONBLOCK);

    return rc;
}

int parse_ulog_message(ebt_ulog_packet_msg_t *msg)
{
    int i = 0;
    char ipbuf[32] = {0};
    char macbuf[32] = {0};
    char *p = NULL;
    int curr_len = ETH_HLEN;
    struct ethhdr *ehdr;
    struct protoent *prototype;
    struct iphdr *iph = NULL;

    assert(msg != NULL);
    if ((msg->version != EBT_ULOG_VERSION) || (msg->data_len < curr_len)) 
        return -1;

    ehdr = (struct ethhdr *)msg->data;
    printf("MAC_SRC=%s\n", ether_ntoa((const struct ether_addr *)ehdr->h_source));
    sprintf(macbuf, "%s", ether_ntoa((const struct ether_addr *)ehdr->h_dest));
    printf("MAC_DST=%s\n", macbuf);

    if (ehdr->h_proto == htons(ETH_P_8021Q)) {
        struct vlan_hdr *vlanh = (struct vlan_hdr *)(((char *)ehdr) + curr_len);

        curr_len += VLAN_HLEN;
        if (msg->data_len < curr_len) {
            return -1;
        }
        printf("VLAN_TCI=%d\n", ntohs(vlanh->TCI));
        if (vlanh->encap == htons(ETH_P_IP)) {
            iph = (struct iphdr *)(vlanh + 1);
        }
    } else if (ehdr->h_proto == htons(ETH_P_IP))
        iph = (struct iphdr *)(((char *)ehdr) + curr_len);

    if (iph == NULL)
        return -1;
    curr_len += sizeof(struct iphdr);
    if (msg->data_len < curr_len)
        return -1;
    curr_len += iph->ihl * 4 - sizeof(struct iphdr);
    if (msg->data_len < curr_len)
        return -1;

    printf("IP_SRC_ADDR=");
    for (i = 0; i < 4; i++) {
        printf("%d%s", ((unsigned char *)&iph->saddr)[i], (i == 3) ? "" : ".");
    }
    printf("\nIP_DEST_ADDR=");
    p = ipbuf; 
    for (i = 0; i < 4; i++) {
        printf("%d%s", ((unsigned char *)&iph->daddr)[i], (i == 3) ? "" : ".");
        sprintf(p, "%d%s", ((unsigned char *)&iph->daddr)[i], (i == 3) ? "" : ".");
        p = ipbuf + strlen(ipbuf); 
    }
    printf("\nIP_PROTOCOL=");
    if (!(prototype = getprotobynumber(iph->protocol))) {
        printf("%d\n", iph->protocol);
    } else {
        struct dhcp_message *dhcp_msg = (struct dhcp_message *)(iph + 1);
    	struct in_addr gin = {0};

        printf("%s\n", prototype->p_name);
        gin.s_addr = dhcp_msg->giaddr;
	if (gin.s_addr != 0) {
            sprintf(ipbuf, "%s", inet_ntoa(gin));
            printf("IP_ADDR = %s\n", ipbuf);
	}
        sprintf(macbuf, "%s", ether_ntoa((const struct ether_addr *)(dhcp_msg->chaddr + 8)));
        printf("HW_ADDR = %s\n", macbuf);
    }
    if (strncmp(ipbuf, "255.255.255.255", 4) != 0) {
        insert(ipbuf, macbuf, msg->outdev);
    }

    return 0;
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
    int rc, grp;
    int ux_sock, max_sock;
    ebt_ulog_packet_msg_t *msg;
    fd_set rset;
    struct timeval tm;

    if (argc < 2)
        usage();
    grp = atoi(argv[1]);
    if ((grp < 1) || (grp > 32))
        usage();
    daemonize();

    rc = init_ulog_socket(grp);
    ux_sock = init_lookup_socket();
    max_sock = (ux_sock > nl_sock) ? ux_sock : nl_sock;
    while (1) {
        FD_ZERO(&rset);
        FD_SET(nl_sock, &rset);
        FD_SET(ux_sock, &rset);
        tm.tv_sec = 30;
        tm.tv_usec = 0;

        rc = select(max_sock+1, &rset, NULL, NULL, &tm);
        if (rc == 0) {
            check_entries();
            continue;
        }

        if (FD_ISSET(nl_sock, &rset)) {
            msg = get_ulog_packet();
            if (msg != NULL)
                rc = parse_ulog_message(msg);
        }
        if (FD_ISSET(ux_sock, &rset)) {
            reply_query(ux_sock);
        }
    }

    return 0;
}
