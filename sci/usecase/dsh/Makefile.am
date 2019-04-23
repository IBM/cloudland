dsh_fe64_SOURCES = dsh_fe.c
edsh_fe64_SOURCES = edsh_fe.cpp
gdsh_fe64_SOURCES = gdsh_fe.cpp
ddsh_fe64_SOURCES = ddsh_fe.c
dsh_be64_SOURCES = dsh_be.c
dsh_fe_SOURCES = dsh_fe.c
edsh_fe_SOURCES = edsh_fe.cpp
gdsh_fe_SOURCES = gdsh_fe.cpp
ddsh_fe_SOURCES = ddsh_fe.c
dsh_be_SOURCES = dsh_be.c
if SYSTEM_AIX
if BIT64
bin_PROGRAMS = dsh_fe64 edsh_fe64 gdsh_fe64 ddsh_fe64 dsh_be64
dsh_fe64_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
dsh_fe64_LDFLAGS = -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
edsh_fe64_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
edsh_fe64_LDFLAGS = -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
gdsh_fe64_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
gdsh_fe64_LDFLAGS = -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
ddsh_fe64_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
ddsh_fe64_LDFLAGS = -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
dsh_be64_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
dsh_be64_LDFLAGS = -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
else
bin_PROGRAMS = dsh_fe edsh_fe gdsh_fe ddsh_fe dsh_be
dsh_fe_CPPFLAGS = -g -O0 -I../../common -I../../include
dsh_fe_LDFLAGS = -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
edsh_fe_CPPFLAGS = -g -O0 -I../../common -I../../include
edsh_fe_LDFLAGS = -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
gdsh_fe_CPPFLAGS = -g -O0 -I../../common -I../../include
gdsh_fe_LDFLAGS = -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
ddsh_fe_CPPFLAGS = -g -O0 -I../../common -I../../include
ddsh_fe_LDFLAGS = -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
dsh_be_CPPFLAGS = -g -O0 -I../../common -I../../include
dsh_be_LDFLAGS = -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
endif
endif
if SYSTEM_DARWIN
bin_PROGRAMS = dsh_fe edsh_fe gdsh_fe ddsh_fe dsh_be
dsh_fe_CPPFLAGS = -I../../common -I../../include
dsh_fe_LDFLAGS = -L../../common -L../../libsci -lcmn -lpthread -lsci
edsh_fe_CPPFLAGS = -I../../common -I../../include
edsh_fe_LDFLAGS = -L../../common -L../../libsci -lcmn -lpthread -lsci
gdsh_fe_CPPFLAGS = -I../../common -I../../include
gdsh_fe_LDFLAGS = -L../../common -L../../libsci -lcmn -lpthread -lsci
ddsh_fe_CPPFLAGS = -I../../common -I../../include
ddsh_fe_LDFLAGS = -L../../common -L../../libsci -lcmn -lpthread -lsci
dsh_be_CPPFLAGS = -I../../common -I../../include
dsh_be_LDFLAGS = -L../../common -L../../libsci -lcmn -lpthread -lsci
endif
if SYSTEM_LINUX
if BIT64
bin_PROGRAMS = dsh_fe64 edsh_fe64 gdsh_fe64 ddsh_fe64 dsh_be64
dsh_fe64_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
dsh_fe64_LDFLAGS = -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
edsh_fe64_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
edsh_fe64_LDFLAGS = -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
gdsh_fe64_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
gdsh_fe64_LDFLAGS = -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
ddsh_fe64_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
ddsh_fe64_LDFLAGS = -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
dsh_be64_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
dsh_be64_LDFLAGS = -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
else
bin_PROGRAMS = dsh_fe edsh_fe gdsh_fe ddsh_fe dsh_be
dsh_fe_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
dsh_fe_LDFLAGS = -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
edsh_fe_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
edsh_fe_LDFLAGS = -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
gdsh_fe_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
gdsh_fe_LDFLAGS = -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
ddsh_fe_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
ddsh_fe_LDFLAGS = -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
dsh_be_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
dsh_be_LDFLAGS = -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
endif
endif
