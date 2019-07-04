CC=gcc
CFLAGS=-g -Wall -Wunused -Werror
SOURCES=getip.c dhcp.c lookup.c

all: getipaddr askip

getipaddr: $(SOURCES)
	$(CC) $(CFLAGS) -o $@ $(SOURCES)

askip: askip.c
	$(CC) $(CFLAGS) -o $@ $<

clean:
	rm -f *.o getipaddr askip
