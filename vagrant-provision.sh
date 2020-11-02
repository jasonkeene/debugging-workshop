#!/bin/bash

set -Eeuo pipefail

# set to true to debug script output
DEBUG=false

function install_go {
    echo Installing Go

    go_version=1.15.3
    go_sum=010a88df924a81ec21b293b5da8f9b11c176d27c0ee3962dc1738d2352d3c02d
    echo -e "\tinstalling version $go_version"
    mkdir /godl
    pushd /godl > /dev/null
        wget -q -O go.tar.gz https://golang.org/dl/go$go_version.linux-amd64.tar.gz
        sha256sum --quiet -c <(echo "$go_sum go.tar.gz")
        tar -C /usr/local -xzf go.tar.gz
    popd > /dev/null
    rm -r /godl

    echo -e "\tsetting up path env vars"
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/vagrant/.profile
    echo 'export PATH=$PATH:/home/vagrant/go/bin' >> /home/vagrant/.profile
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.profile
    echo 'export PATH=$PATH:/root/go/bin' >> /root/.profile

    echo Done Installing Go
}

function install_delve {
    echo Installing Delve

    echo -e "\tinstalling for root user"
    go get github.com/go-delve/delve/cmd/dlv

    echo -e "\tinstalling for vagrant user"
    sudo -iu vagrant go get github.com/go-delve/delve/cmd/dlv

    echo Done Installing Delve
}

function install_docker {
    echo Installing Docker

    echo -e "\tinstalling apt repository"
    apt-get update -qq
    apt-get install -qq --yes \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg-agent \
        software-properties-common >&2
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - >&2 
    add-apt-repository \
       "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
       $(lsb_release -cs) \
       stable" >&2

    echo -e "\tinstalling docker ce"
    apt-get update -qq
    apt-get install -qq --yes \
        docker-ce \
        docker-ce-cli \
        containerd.io >&2

    echo -e "\tadding vagrant user to docker group"
    usermod -aG docker vagrant

    echo -e "\tpulling go image"
    docker pull -q golang:1.15 >&2 &

    echo Done Installing Docker
}

function install_rr {
    echo Installing rr

    mkdir /rrdl
    pushd /rrdl > /dev/null
        wget -q https://github.com/mozilla/rr/releases/download/5.3.0/rr-5.3.0-Linux-$(uname -m).deb
        dpkg -i rr-5.3.0-Linux-$(uname -m).deb >&2
    popd > /dev/null
    rm -r /rrdl

    echo Done Installing rr
}

function install_bpftrace {
    echo Installing BCC/bpftrace

    echo -e "\tinstalling dependencies"
    apt-get update -qq
    apt-get install -qq --yes bison cmake flex g++ git libelf-dev zlib1g-dev \
        libfl-dev systemtap-sdt-dev systemtap-sdt-dev binutils-dev \
        llvm-7-dev llvm-7-runtime libclang-7-dev clang-7 build-essential \
        python libedit-dev luajit luajit-5.1-dev >&2

    mkdir /bpfdl
    pushd /bpfdl > /dev/null
        echo -e "\tcloning repos"
        git clone -n -q --recursive https://github.com/iovisor/bcc
        pushd bcc > /dev/null
            git checkout -q v0.16.0
        popd > /dev/null
        git clone -n -q --recursive https://github.com/iovisor/bpftrace
        pushd bpftrace > /dev/null
            git checkout -q v0.11.1
        popd > /dev/null

        echo -e "\tbuilding bcc"
        mkdir bcc/build
        pushd bcc/build > /dev/null
            cmake -DCMAKE_INSTALL_PREFIX=/usr -DCMAKE_BUILD_TYPE=Debug .. >&2
            make -j2 >&2
            make install >&2
        popd > /dev/null

        echo -e "\tbuilding bpftrace"
        mkdir bpftrace/build
        pushd bpftrace/build > /dev/null
            cmake -DCMAKE_BUILD_TYPE=Debug .. >&2
            make -j2 >&2
            make install >&2
        popd > /dev/null
    popd > /dev/null
    rm -r /bpfdl

    echo Done Installing BCC/bpftrace
}

function install_gdb {
    echo Installing gdb

    apt-get update -qq
    apt-get install -qq --yes gdb >&2

    echo Done Installing gdb
}

function install_jq {
    apt-get install -qq --yes jq >&2
}

function main {
    install_go
    install_delve
    install_docker
    install_rr
    install_bpftrace
    install_gdb
    install_jq
}

[ "$DEBUG" = "false" ] && exec 2>/dev/null
main
wait
