FROM centos:centos7

ARG BRANCH
ARG REPO

WORKDIR /

RUN yum install -y sudo git vim && \
    useradd cland && \
    echo 'cland ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers.d/cland

USER cland

RUN cd /opt && \
    sudo git clone --branch=$BRANCH https://github.com/$REPO.git /opt/cloudland && \
    cd /opt/cloudland/deploy && \
    bash ./allinone.sh

EXPOSE 22
EXPOSE 80
EXPOSE 443
