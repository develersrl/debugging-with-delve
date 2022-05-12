#!/bin/bash

set -Eeuo pipefail

# set to true to debug script output
DEBUG=false

function install_build_essential {
    echo "Installing build-essential"

    apt-get update -qq
    apt install -qq -y build-essential

    echo "Done Installing build-essential"
}

function install_go {
    echo "Installing Go"

    go_version=1.18.2
    go_sum=e54bec97a1a5d230fc2f9ad0880fcbabb5888f30ed9666eca4a91c5a32e86cbc
    echo -e "\tinstalling version $go_version"
    mkdir /godl
    pushd /godl > /dev/null
        wget -q -O go.tar.gz https://go.dev/dl/go$go_version.linux-amd64.tar.gz
        sha256sum --quiet -c <(echo "$go_sum go.tar.gz")
        tar -C /usr/local -xzf go.tar.gz
    popd > /dev/null
    rm -r /godl

    echo -e "\tsetting up path env vars"
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/vagrant/.profile
    echo 'export PATH=$PATH:/home/vagrant/go/bin' >> /home/vagrant/.profile
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.bashrc
    echo 'export PATH=$PATH:/root/go/bin' >> /root/.bashrc

    echo "Done Installing Go"
}

function install_delve {
    echo "Installing Delve"

    echo -e "\tinstalling for root user"
    go install github.com/go-delve/delve/cmd/dlv@v1.8.3

    echo -e "\tinstalling for vagrant user"
    sudo -iu vagrant go install github.com/go-delve/delve/cmd/dlv@v1.8.3

    echo "Done Installing Delve"
}

function install_systemd_coredump {
    echo "Installing systemd_coredump"

    apt-get update -qq
    apt install -qq -y systemd-coredump

    echo "Done installing systemd_coredump"
}

function install_bpfcc_tools {
    echo "Installing bpfcc-tools"

    add-apt-repository -y universe
    apt-get update -qq
    apt install -qq -y bpfcc-tools

    echo "Done Installing bpfcc-tools"
}

function install_gdb {
    echo "Installing gdb"

    apt-get update -qq
    apt-get install -qq -y gdb

    echo "Done Installing gdb"
}

function configure_ptrace {
    echo "Configuring ptrace yama setting"

    echo 0 > /proc/sys/kernel/yama/ptrace_scope

    echo "Done Configuring ptrace yama setting"
}

function install_docker {
    echo "Installing Docker"

    echo -e "\tinstalling apt repository"
    apt-get update -qq
    apt-get install -qq --yes \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg \
        lsb-release
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
    add-apt-repository \
       "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
       $(lsb_release -cs) \
       stable"

    echo -e "\tinstalling docker ce"
    apt-get update -qq
    apt-get install -qq -y \
        docker-ce \
        docker-ce-cli \
        containerd.io

    echo -e "\tadding vagrant user to docker group"
    usermod -aG docker vagrant

    echo -e "\tpulling go image"
    docker pull -q golang:1.18

    echo "Done Installing Docker"
}

function main {
    install_build_essential
    install_go
    install_delve
    install_systemd_coredump
    install_bpfcc_tools
    install_gdb
    configure_ptrace
    install_docker
}

[ "$DEBUG" = "false" ] && exec 2>/dev/null
main
wait
