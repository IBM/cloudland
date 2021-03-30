#!/bin/bash

set -ex

# Check root
if [[ `whoami` != "root" ]]; then
    echo "Not root"
    exit -1
fi

# Install tools
yum groupinstall -y "Development Tools"
yum install epel-release
yum install -y git jq
grep -q 'release 7' /etc/redhat-release
if [ $? -eq 0 ]; then
    wget https://github.com/Kitware/CMake/releases/download/v3.20.0/cmake-3.20.0-linux-x86_64.sh -O /root/cmake.sh
    chmod +x /root/cmake.sh
    /root/cmake.sh --skip-license --prefix=/usr/local
    export PATH=/usr/local/bin:$PATH
    yum -y install centos-release-scl
    yum -y install devtoolset-9-gcc devtoolset-9-gcc-c++ devtoolset-9-binutils
    source /opt/rh/devtoolset-9/enable
fi

# Get release tag from input
release_tag="latest"
if [[ $# -eq 1 ]]; then
    release_tag=$1
fi

# Download or not ?
download=1
latest_release=$(curl -s https://api.github.com/repos/grpc/grpc/releases/latest | jq -r .tag_name)
if [[ -e "/root/cloudland-grpc/release_tag" ]]; then
    current_release=$(cat /root/cloudland-grpc/release_tag)
    if [[ "$current_release" = "$release_tag" || "$current_release" = "$latest_release" ]]; then
        download=0
    fi
fi

if [[ $download -eq 1 ]]; then
    # grpc target folder
    rm -rf /root/cloudland-grpc
    mkdir -p /root/cloudland-grpc

    # Clear grpc
    rm -rf /root/grpc
    
    # Download source code
    cd /root
    if [[ "$release_tag" = "latest" ]]; then
        release_tag=$(curl -s https://api.github.com/repos/grpc/grpc/releases/latest | jq -r .tag_name)
        echo "$release_tag" > /root/cloudland-grpc/release_tag
    fi
    git clone -b "$release_tag" https://github.com/grpc/grpc

    # Update submodule
    cd /root/grpc
    git submodule update --init

    # Get and save commitID
    commitID=$(git rev-parse --short HEAD 2>/dev/null)
    if [ "$commitID" = "" ]; then
        commitID="NaN"
    fi
    echo "$commitID" > /root/cloudland-grpc/commit
fi

pushd "/root/grpc"

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
tar xzf grpc-${commitID}.tar.gz -C /

popd
