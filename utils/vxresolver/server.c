#include <sys/un.h>
#include <unistd.h>
#include <fcntl.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <netinet/ether.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <sqlite3.h>
#include <libpq-fe.h>

#include "server.h"

#define SOCKET_FILE "/var/run/vxhelper.sock"

struct db_desc {
	int db_type;
	PGconn *conn;
	sqlite3 *dbf;
};

struct fdb_addr {
	in_addr_t inner_ip;
	in_addr_t outer_ip;
	u_int8_t inner_mac[ETH_ALEN];
	u_int8_t outer_mac[ETH_ALEN];
};

int init_unix_socket()
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

int init_udp_socket(in_addr_t addr, int port)
{
	int usock, rc, val;
	unsigned int yes = 1;
	struct sockaddr_in srvaddr;

	usock = socket(AF_INET, SOCK_DGRAM, 0);
	memset((char *) &srvaddr, 0, sizeof(srvaddr));
	srvaddr.sin_family = AF_INET;
	srvaddr.sin_port = htons(port);
	srvaddr.sin_addr.s_addr = addr;
	if (port == 0) {
		rc = setsockopt(usock, SOL_IP, IP_TRANSPARENT, &yes, sizeof(yes));
	}
	rc = bind(usock, (struct sockaddr*)&srvaddr, sizeof(srvaddr));
	if (rc != 0)
		return rc;
	val = fcntl(usock, F_GETFL, 0);
	if (val != -1)
		rc = fcntl(usock, F_SETFL, val | O_NONBLOCK);

	return usock;
}

int do_cmd(char *command)
{
	if (strncasecmp(command, "add", 3) == 0) {
		char *pip = command + 4;
	} else if (strncasecmp(command, "del", 3) == 0) {
		char *pip = command + 4;
	} else {
		printf("Unknown command %s!", command);
		return -1;
	}

	return 0;
}

static int callback(void *data, int argc, char **argv, char **azColName)
{
	int rc = 0;
	int i, j;
	struct fdb_addr *fdb = (struct fdb_addr *)data;

	if (argc < 1) {
		return -1;
	}

	for (i = 0; i < argc; i++) {
		if (strcmp(azColName[i], "inner_mac") == 0) {
			int values[ETH_ALEN] = {0};
			int n = sscanf(argv[i], "%x:%x:%x:%x:%x:%x%*c", &values[0], &values[1], &values[2], &values[3], &values[4], &values[5]);
			if (n != 6) {
				rc = -1;
				break;
			}
			for (j = 0; j < ETH_ALEN; j++) {
				fdb->inner_mac[j] = (uint8_t)values[j];
			}
		} else if (strcmp(azColName[i], "outer_ip") == 0) {
			fdb->outer_ip = inet_addr(argv[i]);
		}
	}

	return rc;
}

int sql_exec(char *sql, struct db_desc *db, void *data)
{
	int rc;

	if (db->db_type == SQLITE3) {
		char *zErrMsg = 0;
		rc = sqlite3_exec(db->dbf, sql, callback, data, &zErrMsg);
		if (rc != SQLITE_OK) {
			fprintf(stderr, "SQL error: %s\n", zErrMsg);
			sqlite3_free(zErrMsg);
		}
	} else if (db->db_type == POSTGRES) {
		PGresult *res = PQexec(db->conn, sql);
		if (PQresultStatus(res) != PGRES_TUPLES_OK) {
			fprintf(stderr, "No data retrieved\n");        
			PQclear(res);
			rc = -1;
		}  
	}

	return rc;
}

int resolve(int usock, int dst_port, struct db_desc *db)
{
	int cnt, rc, vni;
	struct sockaddr_in from_addr = {0};
	socklen_t fromlen = sizeof(from_addr);
	char buf[1450] = {0};
	char sql[512] = {0};
	struct sockaddr_in to_addr = {0};
	int ssock = -1;
	struct vxlanhdr *vxh = NULL;
	struct ether_header *ether = NULL;
	struct arphdr *arph = NULL;
	struct  ether_arp *arp = NULL;
	struct in_addr r_ip;
	struct fdb_addr fdb = {0};

	cnt = recvfrom(usock, buf, sizeof(buf), 0, (struct sockaddr *)&from_addr, &fromlen);
	vxh = (struct vxlanhdr *)buf;
	ether = (struct ether_header *)(vxh + 1);
	if (ether->ether_type != ntohs(ETHERTYPE_ARP)) {
		return 0;
	}
	arp = (struct ether_arp *)(ether + 1);
	arph = &arp->ea_hdr;
	if (arph->ar_op != htons(ARPOP_REQUEST)) {
		return 0;
	}

	vni = ntohl(vxh->vx_vni) >> 8;
	memcpy(&r_ip, arp->arp_tpa, sizeof(r_ip));
	sprintf(sql, "select * from vtep where vni=%d and inner_ip='%s' and status='active'", vni, inet_ntoa(r_ip));
	rc = sql_exec(sql, db, &fdb);
	if (rc != 0) {
		return -1;
	}

	if (fdb.outer_ip == inet_addr("127.0.0.1")) {
		return 0;
	}

	memcpy(ether->ether_dhost, ether->ether_shost, sizeof(ether->ether_dhost));
	memcpy(ether->ether_shost, fdb.inner_mac, sizeof(ether->ether_shost));
	arph->ar_op = htons(ARPOP_REPLY);
	memcpy(arp->arp_tha, arp->arp_sha, sizeof(arp->arp_tha));
	memcpy(arp->arp_tpa, arp->arp_spa, sizeof(arp->arp_tpa));
	memcpy(arp->arp_sha, fdb.inner_mac, sizeof(arp->arp_sha));
	memcpy(arp->arp_spa, &r_ip, sizeof(arp->arp_spa));

	ssock = init_udp_socket(fdb.outer_ip, 0);
	to_addr.sin_family = AF_INET;
	to_addr.sin_port = htons(dst_port);
	to_addr.sin_addr.s_addr = from_addr.sin_addr.s_addr;
	cnt = sendto(ssock, buf, cnt, 0, (struct sockaddr *) &to_addr, sizeof(to_addr));
	close(ssock);

	return 0;
}

int db_open(int db_type, char * db_url, struct db_desc *db)
{
	if (db_type == POSTGRES) {
		PGconn *conn = PQconnectdb(db_url);
		if (PQstatus(conn) == CONNECTION_BAD) {
			fprintf(stderr, "Connection to database failed: %s\n", PQerrorMessage(conn));
			PQfinish(conn);    
			exit(-1);
		}
		db->conn = conn;
	} else if (db_type == SQLITE3) {
		sqlite3 *dbf;
		int rc = sqlite3_open(db_url, &dbf);
		if (rc != SQLITE_OK) {
			fprintf(stderr, "Can't open database: %s\n", sqlite3_errmsg(dbf));
			exit(-1);
		}
		db->dbf = dbf;
	}

	return 0;
}

int db_close(struct db_desc *db)
{
	if (db->db_type == POSTGRES) {
		PQfinish(db->conn);
	} else if (db->db_type == SQLITE3) {
		sqlite3_close(db->dbf);
	}

	return 0;
}

int working(int vx_port, int dst_port, int db_type, char *db_url)
{
	int n = 0;
	int ux_sock = -1;
	int usock = -1;
	int maxfd = 0;
	fd_set rset;
	char cmdbuf[VLAN_CMD_LEN] = {0};
	struct db_desc db = {0};

	ux_sock = init_unix_socket();
	usock = init_udp_socket(htonl(INADDR_ANY), vx_port);
	db.db_type = db_type;
	db_open(db_type, db_url, &db);

	while (1) {
		FD_ZERO(&rset);
		FD_SET(ux_sock, &rset);
		FD_SET(usock, &rset);
		maxfd = (maxfd > usock) ? maxfd : usock;
		n = select(maxfd+1, &rset, NULL, NULL, NULL);
		if (n <= 0)
			continue;

		if (FD_ISSET(usock, &rset)) {
			n = resolve(usock, dst_port, &db);
		}
		if (FD_ISSET(ux_sock, &rset)) {
			n = recvfrom(ux_sock, cmdbuf, sizeof(cmdbuf), 0, NULL, NULL);
			do_cmd(cmdbuf);
		}
	}
	db_close(&db);

	return 0;
}
