#!/bin/bash

set -ex

# Check root
if [[ `whoami` != "root" ]]; then
    echo "Not root"
    exit -1
fi

cd /root/grpc

# Get and save commitID
commitID=$(git rev-parse --short HEAD 2>/dev/null)
if [ "$commitID" = "" ]; then
    commitID="NaN"
fi
echo "$commitID" > /root/cloudland-grpc/commit

# Install openssl to replace boringssl
yum install -y openssl-devel

# Build and install absl
mkdir -p "third_party/abseil-cpp/cmake/build"
pushd "third_party/abseil-cpp/cmake/build"
cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_POSITION_INDEPENDENT_CODE=TRUE ../..
make -j4 install DESTDIR=/root/cloudland-grpc
popd

# Build and install c-ares
mkdir -p "third_party/cares/cares/cmake/build"
pushd "third_party/cares/cares/cmake/build"
cmake -DCMAKE_BUILD_TYPE=Release ../..
make -j4 install DESTDIR=/root/cloudland-grpc
popd

# Build and install protobuf
mkdir -p "third_party/protobuf/cmake/build"
pushd "third_party/protobuf/cmake/build"
cmake -Dprotobuf_BUILD_TESTS=OFF -DCMAKE_BUILD_TYPE=Release ..
make -j4 install DESTDIR=/root/cloudland-grpc
popd

# Build and install re2
mkdir -p "third_party/re2/cmake/build"
pushd "third_party/re2/cmake/build"
cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_POSITION_INDEPENDENT_CODE=TRUE ../..
make -j4 install DESTDIR=/root/cloudland-grpc
popd

# Build and install zlib
mkdir -p "third_party/zlib/cmake/build"
pushd "third_party/zlib/cmake/build"
cmake -DCMAKE_BUILD_TYPE=Release ../..
make -j4 install DESTDIR=/root/cloudland-grpc
popd

# Install gRPC
mkdir -p "cmake/build"
pushd "cmake/build"
cmake \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_PREFIX_PATH="/root/cloudland-grpc/usr/local" \
  -DgRPC_INSTALL=ON \
  -DgRPC_BUILD_TESTS=OFF \
  -DgRPC_CARES_PROVIDER=package \
  -DgRPC_ABSL_PROVIDER=package \
  -DgRPC_PROTOBUF_PROVIDER=package \
  -DgRPC_RE2_PROVIDER=package \
  -DgRPC_SSL_PROVIDER=package \
  -DgRPC_ZLIB_PROVIDER=package \
  ../..
make -j4 install DESTDIR=/root/cloudland-grpc
popd

# Go to DESTDIR
cd /root/cloudland-grpc

# Get commitID
commitID=$(cat commit)

# Package grpc
tar czf grpc-${commitID}.tar.gz usr

# Install grpc to /usr/local
# tar xzf grpc-${commitID}.tar.gz -C /
