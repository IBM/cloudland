AUTOMAKE_OPTIONS=foreign
SOURCES = \
    allocator.cpp    filterproc.cpp   message.cpp    purifierproc.cpp  routinglist.cpp \
    api.cpp          eventntf.cpp    handlerproc.cpp  observer.cpp   queue.cpp         \
    ctrlblock.cpp    filter.cpp      initializer.cpp  readerproc.cpp    topology.cpp launcher.cpp \
    dgroup.cpp       filterlist.cpp  listener.cpp     processor.cpp  routerproc.cpp    writerproc.cpp \
    embedagent.cpp   privatedata.cpp bemap.cpp

noinst_HEADERS = \
    allocator.hpp  filterproc.hpp listener.hpp processor.hpp routerproc.hpp writerproc.hpp \
    atomic.hpp      eventntf.hpp    general.hpp      message.hpp   purifierproc.hpp  routinglist.hpp \
    ctrlblock.hpp   filter.hpp      handlerproc.hpp  observer.hpp  queue.hpp   embedagent.hpp       \
    dgroup.hpp      filterlist.hpp  initializer.hpp  readerproc.hpp    topology.hpp privatedata.hpp bemap.hpp

include_HEADERS = ../include/sci.h

OBJECTS = $(addsuffix .o, $(basename $(SOURCES)))
OBJECTS64 = $(addsuffix .64o, $(basename $(SOURCES)))

if SYSTEM_AIX
lib_LIBRARIES = libsci_r.a

SCI_CPPFLAGS = -g -O0 -DREENTRANT -D_SCI_DEBUG -I../include -I../common -I../libsci -I../ -I../secmod -qmkshrobj 
libsci_r.a : libsci_r.o libsci64_r.o
	$(AR) -X32_64 -rv $@ $^
libsci_r.o : $(OBJECTS)
	$(CXX) -q32 $(SCI_CPPFLAGS) -o $@ $^ -L../common -lcmn
libsci64_r.o : $(OBJECTS64)
	$(CXX) -q64 $(SCI_CPPFLAGS) -o $@ $^ -L../common -lcmn64

%.o: %.cpp
	$(CXX) -q32 -c $(SCI_CPPFLAGS) -o $@ $< 

%.64o: %.cpp
	$(CXX) -q64 -c $(SCI_CPPFLAGS) -o $@ $< -D__64BIT__

endif
if SYSTEM_LINUX
lib_LTLIBRARIES=libsci.la
libsci_la_SOURCES = $(SOURCES)
libsci_la_CPPFLAGS = -g -O0 -I../include -I../common -I../secmod
if BIT64
libsci_la_CPPFLAGS += -D__64BIT__
endif
libsci_la_LDFLAGS= -L../common -lcmn -lrt -ldl
endif
if SYSTEM_DARWIN
lib_LTLIBRARIES=libsci.la
libsci_la_SOURCES = $(SOURCES)
libsci_la_CPPFLAGS = -I../include -I../common -I../secmod
libsci_la_LDFLAGS= -L../common -lcmn
endif
