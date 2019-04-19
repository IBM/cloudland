#!/bin/bash
if uname | grep Linux &> /dev/null; then
    for pkg in zip unzip build-essential; do
        dpkg -s $pkg &> /dev/null || apt-get install -qq $pkg
    done
fi
make cladmin
rm -f *.zip
zip -r scripts.zip scripts
mkdir -p build/usr/local/bin
cat $GOPATH/bin/cladmin scripts.zip > build/usr/local/bin/cladmin
chmod a+x build/usr/local/bin/cladmin
tar Ccfz build cladmin.tgz usr
cat > cladmin.sum <<EOF
[cladmin]
version = "$(git describe --always --tags)"
sha1sum = "$(shasum cladmin.tgz | cut -c 1-40)"
EOF
tar cfz cladmin-deploy.tgz deploy.sh start.sh
cat > cladmin-deploy.sum <<EOF
[cladmin-deploy]
version = "$(git describe --always --tags)"
sha1sum = "$(shasum cladmin-deploy.tgz | cut -c 1-40)"
deploy = true
EOF
