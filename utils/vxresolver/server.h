#ifndef _SERVER_H_
#define _SERVER_H_

#define MIP_MAX_LEN 32
#define VLAN_CMD_LEN 32

#define VXH_PORT 8896
#define VXLAN_PORT 8472
#define IP_MAX_LEN 64
#define MAXFD 64

enum {
    SQLITE3,
    POSTGRES
};

struct vlan_peer {
	char addr[IP_MAX_LEN];
	struct vlan_peer *next;
};

struct vxlanhdr {
    unsigned int vx_flags;
    unsigned int vx_vni;
};

extern int init_unix_socket();
extern int working(int vx_port, int dst_port, int db_type, char *db_url);

#endif
