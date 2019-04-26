if SYSTEM_AIX
noinst_LIBRARIES = libcmn.a libcmn64.a
else
noinst_LIBRARIES = libcmn.a
endif
SOURCES = \
    envvar.cpp     group.cpp        log.cpp     range.cpp   sshfunc.cpp  thread.cpp \
    exception.cpp  ipconverter.cpp  packer.cpp  socket.cpp  stream.cpp   tools.cpp

ARFLAGS = -rv
libcmn_a_SOURCES = $(SOURCES)
libcmn_a_CPPFLAGS = -I../secmod -fPIC

noinst_HEADERS = \
    envvar.hpp     group.hpp        log.hpp     range.hpp   sshfunc.hpp  thread.hpp \
    exception.hpp  ipconverter.hpp  packer.hpp  socket.hpp  stream.hpp   tools.hpp 

if SYSTEM_AIX
ARFLAGS += -X32_64
libcmn64_a_SOURCES = $(SOURCES)
libcmn64_a_CPPFLAGS = -q64 -I../secmod -fPIC -D__64BIT__
else
if BIT64
libcmn_a_CPPFLAGS += -D__64BIT__
endif
endif
