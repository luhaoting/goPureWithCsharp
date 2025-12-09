# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure("2") do |config|
  # The most common configuration options are documented and commented below.
  # For a complete reference, please see the online documentation at
  # https://docs.vagrantup.com.

  # Every Vagrant development environment requires a box. You can search for
  # boxes at https://vagrantcloud.com/search.
  config.vm.box = "alvistack/ubuntu-24.04"
  config.vm.box_version = "20251010.1.3"

    # 配置虚拟机资源
  config.vm.provider "virtualbox" do |vb|
    vb.memory = "4096"
    vb.cpus = 4
    vb.name = "golang-dev-2"
  end

  # 将 SSH 的主机端口映射为 12222（在本机通过 127.0.0.1:12222 访问）
  # 这允许使用 `ssh -p 12222 vagrant@127.0.0.1` 或 `vagrant ssh -- -p 12222` 连接
  config.vm.network "forwarded_port", guest: 22, host: 12222, host_ip: "127.0.0.1", id: "ssh", auto_correct: true

  # Disable automatic box update checking. If you disable this, then
  # boxes will only be checked for updates when the user runs
  # `vagrant box outdated`. This is not recommended.
  # config.vm.box_check_update = false

  # Create a forwarded port mapping which allows access to a specific port
  # within the machine from a port on the host machine. In the example below,
  # accessing "localhost:8080" will access port 80 on the guest machine.
  # NOTE: This will enable public access to the opened port
  # config.vm.network "forwarded_port", guest: 80, host: 8080

  # Create a forwarded port mapping which allows access to a specific port
  # within the machine from a port on the host machine and only allow access
  # via 127.0.0.1 to disable public access
#   config.vm.network "forwarded_port", guest: 7001, host: 7001, host_ip: "127.0.0.1"

  # Create a private network, which allows host-only access to the machine
  # using a specific IP.
  # config.vm.network "private_network", ip: "192.168.33.10"

  # Create a public network, which generally matched to bridged network.
  # Bridged networks make the machine appear as another physical device on
  # your network.
  # config.vm.network "public_network"

  # Share an additional folder to the guest VM. The first argument is
  # the path on the host to the actual folder. The second argument is
  # the path on the guest to mount the folder. And the optional third
  # argument is a set of non-required options.
  # config.vm.synced_folder "../data", "/vagrant_data"

  # Disable the default share of the current code directory. Doing this
  # provides improved isolation between the vagrant box and your host
  # by making sure your Vagrantfile isn't accessible to the vagrant box.
  # If you use this you may want to enable additional shared subfolders as
  # shown above.
  # config.vm.synced_folder ".", "/vagrant", disabled: true

  # 共享文件夹
  config.vm.synced_folder ".", "/home/vagrant/workspace"
  
  # 自动安装 Golang
  config.vm.provision "shell", inline: <<-SHELL
    # 更新系统
    apt-get update
    
    # 安装基础工具
    apt-get install -y build-essential git curl wget
    
    # 安装 Golang (最新稳定版)
    GO_VERSION="1.25.2"
    wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm go${GO_VERSION}.linux-amd64.tar.gz
    
    # 配置环境变量
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/vagrant/.bashrc
    echo 'export GOPATH=/home/vagrant/go' >> /home/vagrant/.bashrc
    echo 'export PATH=$PATH:$GOPATH/bin' >> /home/vagrant/.bashrc
    
    # 创建 GOPATH 目录
    mkdir -p /home/vagrant/go/{bin,src,pkg}
    chown -R vagrant:vagrant /home/vagrant/go
    
    # 验证安装
    /usr/local/go/bin/go version 
  SHELL
end
