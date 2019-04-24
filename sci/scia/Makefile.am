scia_SOURCES = scia.cpp
scia64_SOURCES = scia.cpp
if SYSTEM_AIX
if BIT64
bin_PROGRAMS = scia64 scia
else
bin_PROGRAMS = scia
endif
scia_CPPFLAGS = -g -O0 -I../common -I../include
scia_LDFLAGS = -L../common -lcmn64 -lrt -lpthread -ldl
scia64_CPPFLAGS = -g -O0 -q64 -I../common -I../include -D__64BIT__
scia64_LDFLAGS = -q64 -L../common -lcmn64 -lrt -lpthread -ldl
endif
if SYSTEM_DARWIN
bin_PROGRAMS = scia
scia_CPPFLAGS = -I../common -I../include
scia_LDFLAGS = -L../common -lcmn -lpthread -ldl
endif
if SYSTEM_LINUX
if BIT64
bin_PROGRAMS = scia64
scia64_CPPFLAGS = -m64 -g -O0 -I../common -I../include -D__64BIT__
scia64_LDFLAGS = -m64 -L../common -lcmn -lrt -lpthread -ldl
else
bin_PROGRAMS = scia
scia_CPPFLAGS = -m32 -g -O0 -I../common -I../include
scia_LDFLAGS = -m32 -L../common -lcmn -lrt -lpthread -ldl
endif
endif
