CPP = g++
AR = ar
RANLIB = ranlib
CPPFLAGS = -g -Wall -Icommon
LDFLAGS = -Lcommon
ARFLAGS = -rcv

libcmn.a: log.o exception.o
	$(AR) $(ARFLAGS) $@ $^
	$(RANLIB) $@

log.o: log.cpp
	$(CPP) $(CPPFLAGS) -c -o $@ $<

exception.o: exception.cpp
	$(CPP) $(CPPFLAGS) -c -o $@ $<

clean:
	rm -f libcmn.a *.o
