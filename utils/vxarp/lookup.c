#include <sys/types.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <unistd.h>
#include <stdio.h>
#include <fcntl.h>
#include <stdlib.h>
#include <ctype.h>
#include <linux/if.h>

#define SOCKET_FILE "/var/run/get_ip_server.sock"
#define TRX_LEN 32

struct ip_mac_map {
    char ip[TRX_LEN];
    char mac[TRX_LEN];
    char bridge[IFNAMSIZ];
    struct ip_mac_map *prev;
    struct ip_mac_map *next;
};

struct ip_mac_map ip_mac_head = { {0}, {0}, {0}, NULL, NULL };

int init_lookup_socket()
{
    int ux_sock, rc, val;
    struct sockaddr_un srvaddr;

    unlink(SOCKET_FILE);
    ux_sock = socket(AF_UNIX, SOCK_DGRAM, 0);
    srvaddr.sun_family = AF_UNIX;
    strcpy(srvaddr.sun_path, SOCKET_FILE);
    rc = bind(ux_sock, (struct sockaddr*)&srvaddr, sizeof(srvaddr));
    if (rc != 0)
        return rc;
    val = fcntl(ux_sock, F_GETFL, 0);
    if (val != -1)
        rc = fcntl(ux_sock, F_SETFL, val | O_NONBLOCK);

    return ux_sock;
}

int lookup_by_ip(char *ip, char *reply)
{
    struct ip_mac_map *p = NULL;

    for (p = ip_mac_head.next; p != NULL; p = p->next) {
        if (strncmp(p->ip, ip, sizeof(p->ip)) != 0) 
            continue;

        memcpy(reply, p->mac, sizeof(p->mac));
        return 0;
    }

    return -1;
}

int lookup_by_mac(char *mac, char *reply)
{
    struct ip_mac_map *p = NULL;

    for (p = ip_mac_head.next; p != NULL; p = p->next) {
        if (strncmp(p->mac, mac, sizeof(p->mac)) != 0) 
            continue;

        memcpy(reply, p->ip, sizeof(p->ip));
        return 0;
    }

    return -1;
}

int insert(char *ip, char *mac, char *bridge)
{
    struct ip_mac_map *ip_mac = NULL;
    struct ip_mac_map *p = NULL;

    for (p = ip_mac_head.next; p != NULL; p = p->next) {
        if (strncmp(p->mac, mac, sizeof(p->mac)) == 0){
            strncpy(p->ip, ip, sizeof(p->ip));
            return 0;
        }	
    }

    ip_mac = (struct ip_mac_map *)malloc(sizeof(struct ip_mac_map));
    memset(ip_mac, 0, sizeof(struct ip_mac_map));
    strncpy(ip_mac->ip, ip, sizeof(ip_mac->ip));
    strncpy(ip_mac->mac, mac, sizeof(ip_mac->mac));
    strncpy(ip_mac->bridge, bridge, sizeof(ip_mac->bridge));
    ip_mac->next = ip_mac_head.next;
    ip_mac->prev = &ip_mac_head;
    if (ip_mac_head.next != NULL)
        ip_mac_head.next->prev = ip_mac;
    ip_mac_head.next = ip_mac;

    return 0;
}

int reply_query(int sock)
{
    int i, n, ch, rc, clilen;
    char buf[TRX_LEN] = {0};
    char reply[TRX_LEN] = {0};
    char macbuf[TRX_LEN] = {0};
    char *p = macbuf;
    struct sockaddr_un cliaddr;

    clilen = sizeof(cliaddr);
    n = recvfrom(sock, buf, sizeof(buf), 0, (struct sockaddr *)&cliaddr, (socklen_t *)&clilen);
    if (buf[TRX_LEN-1] != '\0')
        return -1;

    if (strchr(buf, ':') != NULL) {
        n = strlen(buf);
        for (i = 0; i < n; i++) {
            ch = buf[i+1];
            if ((buf[i] == '0') && (((ch >= '0') && (ch <= '9')) ||
                        ((ch >= 'A') && (ch <= 'F')) ||
                        ((ch >= 'a') && (ch <= 'f'))))
                continue;
            *p = tolower(buf[i]);
            p++;
        }
        rc = lookup_by_mac(macbuf, reply);
    } else if (strchr(buf, '.') != NULL) {
        rc = lookup_by_ip(buf, reply);
    }
    n = sendto(sock, reply, sizeof(reply), 0, (struct sockaddr *)&cliaddr, sizeof(cliaddr));

    return 0;
}

int check_entries()
{
    int ret = -1;
    char cmd[256] = {0};
    struct ip_mac_map *p = NULL;

    for (p = ip_mac_head.next; p != NULL; p = p->next) {
        sprintf(cmd, "arping -w 2 -f -I %s %s", p->bridge, p->ip);
        ret = system(cmd);
        if (WEXITSTATUS(ret) != 0) {
            p->prev->next = p->next;
            if (p->next != NULL)
                p->next->prev = p->prev;
            free(p);
        }
    }

    return 0;
}
