if SYSTEM_AIX
lib_LIBRARIES = libpsec.a

PSEC_CFLAGS = -g -O0 -D_GNU_SOURCE -U_TRACE -U_DEBUG_IOLIST -U_DEBUG_RDWR -U_DEBUG_ATOM \
              -U_CHECK_MEM_LEAK -U_TRACE_RESPONDER_DELAY -U_TRACE_IO_OP -D_FILE_OFFSET_BITS=64 \
              -D_UTE_TRACE -U_TRACE_IO_FLOW -UUSE_MIO -U_MEM_HW -D_CC_SHARED_LOCK -U_ENABLE_CC_ASSERT -qmkshrobj -I../
libpsec.a : shr_64.o shr.o libpsec_ossh.so libpsec_ossh64.so
	$(AR) -X32_64 -rv $@ shr_64.o shr.o

shr_64.o : psec.64o psec_utils.64o
	$(CC) -q64 $(PSEC_CFLAGS) -o $@ $^

shr.o : psec.o psec_utils.o
	$(CC) -q32 $(PSEC_CFLAGS) -o $@ $^

%.o: %.c 
	$(CC) -q32 -c $(PSEC_CFLAGS) -o $@ $< 

%.64o: %.c
	$(CC) -q64 -c $(PSEC_CFLAGS) -o $@ $< 

libpsec_ossh.so : psec_openssh.c
	$(CC) -q32 -g -O0 $(PSEC_CFLAGS) -o $@ $^ -lcrypto

libpsec_ossh64.so : psec_openssh.c
	$(CC) -q64 -g -O0 $(PSEC_CFLAGS) -o $@ $^ -lcrypto

install-exec-hook:
	cp *.so /usr/lib

clean-hook:
	rm -f *.a *.so

else
lib_LTLIBRARIES = libpsec.la libpsec_ossh.la
libpsec_la_SOURCES = psec.c psec_utils.c 
libpsec_ossh_la_SOURCES = psec_openssh.c
libpsec_la_CFLAGS = $(PSEC_CFLAGS)
libpsec_ossh_la_CFLAGS = $(PSEC_CFLAGS)
install-exec-hook:
	-ln -sf $(libdir)/libpsec.so /usr/`basename $(libdir)`/libpsec.so
	-ln -sf $(libdir)/libpsec_ossh.so /usr/`basename $(libdir)`/libpsec_ossh.so
uninstall-hook:
	-unlink /usr/`basename $(libdir)`/libsci.so
	-unlink /usr/`basename $(libdir)`/libsci_ossh.so
endif
