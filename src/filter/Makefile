CPP = g++
AR = ar
CPPFLAGS = -g -Wall -I../common -I/opt/sci/include -shared -fPIC
LDFLAGS = -L../common -lcmn -L/opt/sci/lib64 -L/opt/sci/lib -lsci 

all: scheduler.so

scheduler.so: scheduler.o rcmanager.o
	$(CPP) $(CPPFLAGS) -o $@ $^ $(LDFLAGS)

scheduler.o: scheduler.cpp rcmanager.o
	$(CPP) $(CPPFLAGS) -c -o $@ $<

rcmanager.o: rcmanager.cpp
	$(CPP) $(CPPFLAGS) -c -o $@ $< $(LDFLAGS)

clean:
	rm -f *.so *.o

install:
	cp *.so /opt/cloudland/lib64
#	cp *.so /opt/cloudland/lib64-1
