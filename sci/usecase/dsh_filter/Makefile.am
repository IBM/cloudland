dsh_fe64_SOURCES = dsh_fe.cpp levenshtein.cpp dsh_packer.cpp dsh_header.hpp dsh_aggregator.hpp
dsh_be64_SOURCES = dsh_be.cpp levenshtein.cpp dsh_packer.cpp dsh_header.hpp dsh_aggregator.hpp
dsh_tests64_SOURCES = dsh_tests.cpp levenshtein.cpp dsh_packer.cpp dsh_header.hpp dsh_aggregator.hpp
dsh_filter64_la_SOURCES = dsh_filter.cpp levenshtein.cpp dsh_packer.cpp dsh_header.hpp dsh_aggregator.hpp
dsh_fe_SOURCES = dsh_fe.cpp levenshtein.cpp dsh_packer.cpp dsh_header.hpp dsh_aggregator.hpp
dsh_be_SOURCES = dsh_be.cpp levenshtein.cpp dsh_packer.cpp dsh_header.hpp dsh_aggregator.hpp
dsh_tests_SOURCES = dsh_tests.cpp levenshtein.cpp dsh_packer.cpp dsh_header.hpp dsh_aggregator.hpp
dsh_filter_la_SOURCES = dsh_filter.cpp levenshtein.cpp dsh_packer.cpp dsh_header.hpp dsh_aggregator.hpp
if SYSTEM_AIX
if BIT64
bin_PROGRAMS = dsh_fe64 dsh_be64 dsh_tests64
pkglib_LTLIBRARIES = dsh_filter64.la
dsh_fe64_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
dsh_fe64_LDFLAGS = -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
dsh_be64_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
dsh_be64_LDFLAGS = -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
dsh_tests64_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
dsh_tests64_LDFLAGS = -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
dsh_filter64_la_CPPFLAGS = -g -O0 -q64 -I../../common -I../../include -D__64BIT__
dsh_filter64_la_LDFLAGS = -module -q64 -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
else
bin_PROGRAMS = dsh_fe dsh_be dsh_tests
pkglib_LTLIBRARIES = dsh_filter.la
dsh_fe_CPPFLAGS = -g -O0 -I../../common -I../../include
dsh_fe_LDFLAGS = -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
dsh_be_CPPFLAGS = -g -O0 -I../../common -I../../include
dsh_be_LDFLAGS = -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
dsh_tests_CPPFLAGS = -g -O0 -I../../common -I../../include
dsh_tests_LDFLAGS = -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
dsh_filter_la_CPPFLAGS = -g -O0 -I../../common -I../../include
dsh_filter_la_LDFLAGS = -module -L../../common -L../../libsci -lcmn64 -lrt -lpthread -lsci
endif
endif
if SYSTEM_DARWIN
bin_PROGRAMS = dsh_fe dsh_be dsh_tests
pkglib_LTLIBRARIES = dsh_filter.la
dsh_fe_CPPFLAGS = -I../../common -I../../include
dsh_fe_LDFLAGS = -L../../common -L../../libsci -lcmn -lpthread -lsci
dsh_be_CPPFLAGS = -I../../common -I../../include
dsh_be_LDFLAGS = -L../../common -L../../libsci -lcmn -lpthread -lsci
dsh_tests_CPPFLAGS = -I../../common -I../../include
dsh_tests_LDFLAGS = -L../../common -L../../libsci -lcmn -lpthread -lsci
dsh_filter_la_CPPFLAGS = -I../../common -I../../include
dsh_filter_la_LDFLAGS = -module -L../../common -L../../libsci -lcmn -lpthread -lsci
endif
if SYSTEM_LINUX
if BIT64
bin_PROGRAMS = dsh_fe64 dsh_be64 dsh_tests64
pkglib_LTLIBRARIES = dsh_filter64.la
dsh_fe64_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
dsh_fe64_LDFLAGS = -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
dsh_be64_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
dsh_be64_LDFLAGS = -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
dsh_tests64_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
dsh_tests64_LDFLAGS = -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
dsh_filter64_la_CPPFLAGS = -m64 -g -O0 -I../../common -I../../include -D__64BIT__
dsh_filter64_la_LDFLAGS = -module -m64 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
else
bin_PROGRAMS = dsh_fe dsh_be dsh_tests
pkglib_LTLIBRARIES = dsh_filter.la
dsh_fe_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
dsh_fe_LDFLAGS = -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
dsh_be_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
dsh_be_LDFLAGS = -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
dsh_tests_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
dsh_tests_LDFLAGS = -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
dsh_filter_la_CPPFLAGS = -m32 -g -O0 -I../../common -I../../include
dsh_filter_la_LDFLAGS = -module -m32 -L../../common -L../../libsci -lcmn -lrt -lpthread -lsci
endif
endif
