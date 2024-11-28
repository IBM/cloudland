# Build from source (for development) and Build RPM package

## Prerequisite
A Linux build server (For s390x, like LinuxOne, please use RedHat Enterprise Linux 8.3+, for x86-64, the OS needs to be 7.5+) with following tools installed:
1. Build tools:
   1. gRPC C++ implementation:
      - Installation path: /usr/local (dafault gRPC path to build RPM package)
   2. make, gcc, etc. to build C/C++ codes
      - For SCI binaries and CloudLand binaries
   3. golang
      - For web/clui
2. git to access GitHub
3. rpmbuild to build RPM package (optional)

## Build binaries and RPM package
```
# Download source code from GitHub to /opt
git clone https://github.com/IBM/cloudland.git /opt/cloudland

# Go to the source code folder
cd /opt/cloudland

# Build grpc library, skip this step if this is built already
./build_grpc.sh

# Build the binaries. Continue the installation if the build server is also the controller.
./build.sh

# (Optional) Build RPM package after building the binaries
./build_rpm.sh <version> <release>
```

## Additional informaiton
1. The architecture of the build server should be the same as the controller and compute nodes where you want to deploy the CloudLand. For example, if you want to deploy CloudLand on LinuxOne, the s390x architecture, the build server should be a s390x-based build server.
2. Building and running CloudLand need gRPC C++ implementation. When building RPM package, build_rpm.sh will try to package the gRPC libraries from /usr/local/ . Please build and install the gRPC C++ implementation to /usr/local (its default path) before building CloudLand.
3. The source code is downloaded to /opt/cloudland which is fixed in current release.
4. After building the code, there will be three folders:
   1. /opt/cloudland
      1. The main part which contains the source code, the built CloudLand binaries (/opt/cloudland/bin) and web/clui (the GO-implemented service)
   2. /opt/sci
      1. The binaries which CloudLand uses to communicate between controller and compute nodes. Refer to https://wiki.eclipse.org/PTP/designs/SCI for more information.
   3. /opt/libvirt-console-proxy
      1. The websockets console proxy. It's downloaded and built from https://github.com/libvirt/libvirt-console-proxy . Check build.sh for more information.
5. The build.sh also downloades noVNC. Check build.sh for more infromation.
6. build_rpm.sh will create /tmp/grpc.tar.gz and /tmp/cloudland.tar.gz and build a RPM package with the two files.
   1. grpc.tar.gz contains the gRPC files from /usr/local
   2. cloudland.tar.gz contains the three folders listed above. They are copied to /tmp, and are packaged after running 'make clean'
   3. The two zipped files are not deleted automatically now.
7. The RPM package can be installed on controller via yum, which will release the two packages, grpc.tar.gz and cloudland.tar.gz, to /tmp. grpc.tar.gz will be unpacked to / so it will be released to /usr/local. cloudland.tar.gz will be unpacked to /opt so there will be three folders: /opt/cloudland, /opt/sci and /opt/libvirt-console-proxy. The installation only unpacks the files. You need to deploy the controller and compute nodes before using CloudLand. See [Installation](Installation.md) for more information.
