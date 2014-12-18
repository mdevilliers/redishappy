# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"

goLangZip = "go1.4.src.tar.gz"

script = <<SCRIPT

add-apt-repository ppa:vbernat/haproxy-1.5
add-apt-repository ppa:bzr/ppa

# add docker key
apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 36A1D7869245C8950F966E92D8576A8BA88D21E9
echo "deb https://get.docker.io/ubuntu docker main" | sudo tee /etc/apt/sources.list.d/docker.list

apt-get -y update > /dev/null
apt-get upgrade > /dev/null

# install dev tools
apt-get install -y git mercurial ruby-dev gcc wget bzr lintian lxc-docker haproxy redis-server

# stop redis-server
# we only need the client tools
service redis-server stop

# enable haproxy
sed -i 's/^ENABLED=.*/ENABLED=1/' /etc/default/haproxy

# install go
mkdir -p /home/vagrant/go
chown -R vagrant:vagrant /home/vagrant/go

wget https://storage.googleapis.com/golang/#{goLangZip}
tar -C /usr/local -xzf #{goLangZip}
export PATH=$PATH:/usr/local/go/bin:/home/vagrant/go/bin
export GOPATH=/home/vagrant/go

echo "export PATH=$PATH:/usr/local/go/bin:/home/vagrant/go/bin" >> /home/vagrant/.profile
echo "export GOPATH=/home/vagrant/go" >> /home/vagrant/.profile

# install fpm
gem install --no-ri --no-rdoc fpm

go get github.com/mdevilliers/redishappy

go get github.com/tools/godep

cd $GOPATH/src/github.com/mdevilliers/redishappy

godep restore

go get golang.org/x/tools/cmd/cover
go get golang.org/x/tools/cmd/vet
go get golang.org/x/tools/cmd/goimports

export _REDISHAPPY_VERSION="1.0.0"
export _REDISHAPPY_PKGVERSION="1"

build/ci.sh
build/release.sh

# dpkg -i build/redishappy-haproxy_${_REDISHAPPY_VERSION}${_REDISHAPPY_PKGVERSION}_amd64.deb
# dpkg -i build/redishappy-consul_${_REDISHAPPY_VERSION}${_REDISHAPPY_PKGVERSION}_amd64.deb

# download docker testing pre-reqs
cd /home/vagrant
git clone https://github.com/mdevilliers/docker-rediscluster.git

docker pull redis
docker pull joshula/redis-sentinel

SCRIPT

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  
  config.vm.box = 'ubuntu/trusty64'

  config.vm.network :private_network, ip: "192.168.0.22"
  config.vm.network :forwarded_port, guest: 8000, host: 8000
  
  config.vm.provider :virtualbox do |vb|
      vb.customize ["modifyvm", :id, "--memory", 2048,  "--cpus", "2"]
  end

  config.vm.provision :shell, inline: script 

end
