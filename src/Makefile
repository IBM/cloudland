HOST_SYSTEM = $(shell uname | cut -f 1 -d_)
SYSTEM ?= $(HOST_SYSTEM)
CPP = g++
CPPFLAGS = -g -Wall -Icommon -Ifilter -I/opt/sci/include -std=c++11 -DCPPHTTPLIB_OPENSSL_SUPPORT
LDFLAGS = -Lcommon -lcmn -ljsoncpp -L/opt/sci/lib64 -L/opt/sci/lib -lsci -pthread -lcrypto -lssl

ifeq ($(SYSTEM),Darwin)
LDFLAGS += -L/usr/local/lib -ldl
else
LDFLAGS += -L/usr/local/lib -Wl,--no-as-needed -Wl,--as-needed -ldl
endif

LD_BACK_FLAGS=-Lcommon -lcmn -L/opt/sci/lib64 -L/opt/sci/lib -lsci -pthread
ARFLAGS = -rv

all: libcmn.a scheduler.so cloudland cloudlet

libcmn.a:
	$(MAKE) -C common || exit 1

scheduler.so:
	$(MAKE) -C filter || exit 1

cloudland: cloudland.o netlayer.o handler.o rpcworker.o
	$(CPP) $(CPPFLAGS) -o $@ $^ $(LDFLAGS)

cloudlet: cloudlet.o
	$(CPP) $(CPPFLAGS) -o $@ $^ $(LD_BACK_FLAGS)

threadpool: threadpool.o
	$(CPP) -g -Wall -std=c++11 -o $@ $^

rpcworker.o: rpcworker.cpp
	$(CPP) $(CPPFLAGS) -c -o $@ $< $(LDFLAGS)

cloudlet.o: cloudlet.cpp
	$(CPP) $(CPPFLAGS) -c -o $@ $< $(LDFLAGS)

handler.o: handler.cpp
	$(CPP) $(CPPFLAGS) -c -o $@ $< $(LDFLAGS)

threadpool.o: threadpool.cpp
	$(CPP) -g -Wall -pthread -std=c++11 -c -o $@ $<

clean:
	rm -f *.o common/*.o common/*.a remotexec/*.o remotexec/*.a remotexec/*.pb.cc remotexec/*.pb.h filter/*.so filter/*.o cloudland cloudlet cloudctl

install:
	mkdir -p /opt/cloudland/bin /opt/cloudland/lib64
	cp cloudland cloudlet /opt/cloudland/bin
	cp filter/*.so /opt/cloudland/lib64/
