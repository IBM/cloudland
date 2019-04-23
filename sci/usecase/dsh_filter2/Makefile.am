dsh_fe64_SOURCES = dsh_fe.c
dsh_be64_SOURCES = dsh_be.c
downfilter64_la_SOURCES = downfilter.c
upfilter64_la_SOURCES = upfilter.c
upfiltera64_la_SOURCES = upfiltera.c
upfilterb64_la_SOURCES = upfilterb.c
dsh_fe_SOURCES = dsh_fe.c
dsh_be_SOURCES = dsh_be.c
downfilter_la_SOURCES = downfilter.c
upfilter_la_SOURCES = upfilter.c
upfiltera_la_SOURCES = upfiltera.c
upfilterb_la_SOURCES = upfilterb.c
if SYSTEM_AIX
if BIT64
bin_PROGRAMS = dsh_fe64 dsh_be64
pkglib_LTLIBRARIES = downfilter64.la upfilter64.la upfiltera64.la upfilterb64.la
dsh_fe64_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
dsh_fe64_LDFLAGS = -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
dsh_be64_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
dsh_be64_LDFLAGS = -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
downfilter64_la_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
downfilter64_la_LDFLAGS = -module -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
upfilter64_la_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
upfilter64_la_LDFLAGS = -module -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
upfiltera64_la_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
upfiltera64_la_LDFLAGS = -module -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
upfilterb64_la_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
upfilterb64_la_LDFLAGS = -module -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
else
bin_PROGRAMS = dsh_fe dsh_be
pkglib_LTLIBRARIES = downfilter.la upfilter.la upfiltera.la upfilterb.la
dsh_fe_CPPFLAGS = -g -O0 -I../../common -I../../include
dsh_fe_LDFLAGS = -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
dsh_be_CPPFLAGS = -g -O0 -I../../common -I../../include
dsh_be_LDFLAGS = -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
downfilter_la_CPPFLAGS = -g -O0 -I../../common -I../../include
downfilter_la_LDFLAGS = -module -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
upfilter_la_CPPFLAGS = -g -O0 -I../../common -I../../include
upfilter_la_LDFLAGS = -module -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
upfiltera_la_CPPFLAGS = -g -O0 -I../../common -I../../include
upfiltera_la_LDFLAGS = -module -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
upfilterb_la_CPPFLAGS = -g -O0 -I../../common -I../../include
upfilterb_la_LDFLAGS = -module -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
endif
endif
if SYSTEM_DARWIN
bin_PROGRAMS = dsh_fe dsh_be
pkglib_LTLIBRARIES = downfilter.la upfilter.la upfiltera.la upfilterb.la
dsh_fe_CPPFLAGS = -I../../common -I../../include
dsh_fe_LDFLAGS = -L../../common -L../../libsci -lcmn -lpthread -lsci
dsh_be_CPPFLAGS = -I../../common -I../../include
dsh_be_LDFLAGS = -L../../common -L../../libsci -lcmn -lpthread -lsci
downfilter_la_CPPFLAGS = -I../../common -I../../include
downfilter_la_LDFLAGS = -module -L../../common -L../../libsci -lcmn -lpthread -lsci
upfilter_la_CPPFLAGS = -I../../common -I../../include
upfilter_la_LDFLAGS = -module -L../../common -L../../libsci -lcmn -lpthread -lsci
upfiltera_la_CPPFLAGS = -I../../common -I../../include
upfiltera_la_LDFLAGS = -module -L../../common -L../../libsci -lcmn -lpthread -lsci
upfilterb_la_CPPFLAGS = -I../../common -I../../include
upfilterb_la_LDFLAGS = -module -L../../common -L../../libsci -lcmn -lpthread -lsci
endif
if SYSTEM_LINUX
if BIT64
bin_PROGRAMS = dsh_fe64 dsh_be64
pkglib_LTLIBRARIES = downfilter64.la upfilter64.la upfiltera64.la upfilterb64.la
dsh_fe64_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
dsh_fe64_LDFLAGS = -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
dsh_be64_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
dsh_be64_LDFLAGS = -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
downfilter64_la_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
downfilter64_la_LDFLAGS = -module -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
upfilter64_la_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
upfilter64_la_LDFLAGS = -module -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
upfiltera64_la_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
upfiltera64_la_LDFLAGS = -module -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
upfilterb64_la_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
upfilterb64_la_LDFLAGS = -module -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
else
bin_PROGRAMS = dsh_fe dsh_be
pkglib_LTLIBRARIES = downfilter.la upfilter.la upfiltera.la upfilterb.la
dsh_fe_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
dsh_fe_LDFLAGS = -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
dsh_be_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
dsh_be_LDFLAGS = -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
downfilter_la_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
downfilter_la_LDFLAGS = -module -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
upfilter_la_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
upfilter_la_LDFLAGS = -module -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
upfiltera_la_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
upfiltera_la_LDFLAGS = -module -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
upfilterb_la_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
upfilterb_la_LDFLAGS = -module -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
endif
endif
