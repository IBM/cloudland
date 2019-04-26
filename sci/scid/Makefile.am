sbin_PROGRAMS = scidv1
scidv1_SOURCES = extlaunch.cpp  extlisten.cpp  locker.cpp  scid.cpp

if SYSTEM_AIX
scidv1_CPPFLAGS = -q64 -I../common -I../secmod
scidv1_LDFLAGS = -q64 -L../common -lcmn64 -lrt -lpthread -ldl
endif
if SYSTEM_LINUX
scidv1_CPPFLAGS = -I../common -I../secmod
scidv1_LDFLAGS = -L../common -lcmn -lrt -lpthread -ldl
endif
if SYSTEM_DARWIN
scidv1_CPPFLAGS = -I../common -I../secmod
scidv1_LDFLAGS = -L../common -lcmn -lpthread -ldl
endif
