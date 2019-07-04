CC = gcc
CFLAGS = -g -Wall -O0
LDFLAGS = -lsqlite3 -lpq
RONN := ronn

all: vxresolver

vxresolver: main.o server.o
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

#vxadmin: vxadmin.o
#	$(CC) $(CFLAGS) -o $@ $^

#vxadmin.o: vxadmin.c
#	$(CC) $(CFLAGS) -c -o $@ $<

main.o: main.c
	$(CC) $(CFLAGS) -c -o $@ $<

mcast.o: mcast.c mcast.h
	$(CC) $(CFLAGS) -c -o $@ $<

server.o: server.c server.h
	$(CC) $(CFLAGS) -c -o $@ $<

install:
	cp vxresolver /opt/cloudland/bin/

clean:
	rm -f vxresolver *.o *.deb

