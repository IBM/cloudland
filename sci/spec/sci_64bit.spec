Summary: SCI 64bit Run Time Library
Name: sci_64bit
Version: 2
Release: 0.0.0
Group: IBM HPC
License: IBM Corp.
Packager: IBM
URL: http://www.ibm.com
Conflicts: N/A
Vendor: IBM Corp.
Source0: N/A
NoSource: 0
AutoReqProv: no

%description
IBM SCI 64bit Run Time Library
This package contains header files, shared libraries and executables.

%undefine  __check_files

%install

%build

%files
# Library, header files, message catalog and README files.
%attr( 755, bin, bin ) %dir /opt/sci

# Header files
%attr( 755, bin, bin ) %dir /opt/sci/include
%attr( 644, bin, bin ) /opt/sci/include/sci.h

# Libraries.
%attr( 755, bin, bin ) %dir /opt/sci/lib64
%attr( 644, bin, bin ) /opt/sci/lib64/libsci.so
%attr( 644, bin, bin ) /opt/sci/lib64/libsci.so.0
%attr( 644, bin, bin ) /opt/sci/lib64/libsci.so.0.0.0
%attr( 644, bin, bin ) /opt/sci/lib64/libpsec.so
%attr( 644, bin, bin ) /opt/sci/lib64/libpsec.so.0
%attr( 644, bin, bin ) /opt/sci/lib64/libpsec.so.0.0.0
%attr( 644, bin, bin ) /opt/sci/lib64/libpsec_ossh.so
%attr( 644, bin, bin ) /opt/sci/lib64/libpsec_ossh.so.0
%attr( 644, bin, bin ) /opt/sci/lib64/libpsec_ossh.so.0.0.0

# Executables
%attr( 755, bin, bin ) %dir /opt/sci/bin
%attr( 755, bin, bin ) /opt/sci/bin/scia64
%attr( 755, bin, bin ) %dir /opt/sci/sbin
%attr( 755, bin, bin ) /opt/sci/sbin/scidv1

# Pre-install script
%pre

# Post-install script
%post
/sbin/ldconfig

# Pre-uninstall script
%preun

# Post-uninstall script
# Run the following ONLY when performing uninstall (rpm -e),
# skip if it is an upgrape (rpm -U).
%postun

