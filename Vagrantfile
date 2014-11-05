# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"

goLangZip = "go1.3.3.linux-amd64.tar.gz"

script = <<SCRIPT

add-apt-repository ppa:vbernat/haproxy-1.5
add-apt-repository ppa:bzr/ppa
apt-get update -y

# install dev tools
apt-get install -y git mercurial ruby-dev gcc wget bzr lintian

# install haproxy 1.5+
apt-get install -y haproxy
sed -i 's/^ENABLED=.*/ENABLED=1/' /etc/default/haproxy

# install go

wget https://storage.googleapis.com/golang/#{goLangZip}
tar -C /usr/local -xzf #{goLangZip}
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/home/vagrant/go

echo "export PATH=$PATH:/usr/local/go/bin" >> /home/vagrant/.profile
echo "export GOPATH=/home/vagrant/go" >> /home/vagrant/.profile

# install fpm
gem install --no-ri --no-rdoc fpm

go get github.com/mdevilliers/redishappy
go get -t -v ./...

go get code.google.com/p/go.tools/cmd/cover
go get code.google.com/p/go.tools/cmd/vet
go get code.google.com/p/go.tools/cmd/goimports

cd $GOPATH/src/github.com/mdevilliers/redishappy

export _REDISHAPPY_VERSION="1.0.0"
export _REDISHAPPY_PKGVERSION="1"

build/ci.sh
build/release.sh

dpkg -i build/redishappy-haproxy_${_REDISHAPPY_VERSION}${_REDISHAPPY_PKGVERSION}_all.deb
# dpkg -i build/redishappy-consul_${_REDISHAPPY_VERSION}${_REDISHAPPY_PKGVERSION}_all.deb

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
