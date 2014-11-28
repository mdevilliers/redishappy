#!/bin/bash
set -e
set -u
set -x

version=${_REDISHAPPY_VERSION:-"1.0.0"}
url="https://github.com/mdevilliers/redishappy"
arch="amd64"
section="misc"
license="Apache Software License 2.0"
package_version=${_REDISHAPPY_PKGVERSION:-"-1"}
origdir="$(pwd)"
workspace="build"
pkgtype=${_PKGTYPE:-"deb"}
builddir="output"
installdir="usr/bin"
logdir="var/log"
configdir="etc"
vendor="mdevilliers"

function makeRedisHAProxyPackage() {

    name=redishappy-haproxy
    description="RedisHappy HAProxy is an automated Redis failover daemon integrating Redis Sentinel with HAProxy"

    cd ${origdir}/${workspace}
    rm -rf ${name}*.{deb,rpm}
    rm -rf ${builddir}

    mkdir -p ${name}/${logdir}/redishappy-haproxy
    mkdir -p ${name}/${installdir}/../share/doc/redishappy-haproxy
    mkdir -p ${name}/${configdir}/redishappy-haproxy
    mkdir -p ${name}/${configdir}/init

    cp ${origdir}/redis-haproxy ${name}/${installdir}/redis-haproxy
    chmod 755 ${name}/${installdir}/redis-haproxy

    cp ${origdir}/${workspace}/configs/redis-haproxy/config.json ${name}/${configdir}/redishappy-haproxy/config.json
    cp ${origdir}/${workspace}/configs/redis-haproxy/haproxy_template.cfg ${name}/${configdir}/redishappy-haproxy/haproxy_template.cfg

    cp ${origdir}/${workspace}/redishappy-haproxy-service ${name}/${configdir}/init/redishappy-haproxy-service.conf

    pushd ${name}

      # rubygem: fpm
    fpm -t ${pkgtype} \
        -n ${name} \
        -v ${version}${package_version} \
        --description "${description}" \
        --url="${url}" \
        -a ${arch} \
        --category ${section} \
        --vendor ${vendor} \
        -m "${USER}@${HOSTNAME}" \
        --license "${license}" \
        --deb-upstart ../redishappy-haproxy-service \
        --prefix=/ \
        -s dir \
	--config-files etc/init \
	--config-files etc/redishappy-haproxy \
	--pre-install ../local-scripts/preinstall-redishappy-haproxy \
	--post-install ../local-scripts/postinstall-redishappy-haproxy \
	--post-uninstall ../local-scripts/postrm-redishappy-haproxy \
	--debug \
        -- .

  mv ${name}*.${pkgtype} ${origdir}/${workspace}/

  popd
}


function makeRedisConsulPackage() {

    name=redishappy-consul
    description="RedisHappy Consul is an automated Redis failover daemon integrating Redis Sentinel with Consul"

    cd ${origdir}/${workspace}
    rm -rf ${name}*.{deb,rpm}
    rm -rf ${builddir}

    mkdir -p ${name}/${logdir}/redishappy-consul
    mkdir -p ${name}/${configdir}/redishappy-consul
    mkdir -p ${name}/${installdir}/../share/doc/redishappy-consul
    mkdir -p ${name}/${configdir}/init

    cp ${origdir}/redis-consul ${name}/${installdir}/redis-consul
    chmod 755 ${name}/${installdir}/redis-consul

    cp ${origdir}/main/redis-consul/config.json ${name}/${configdir}/redishappy-consul/config.json

    cp ${origdir}/${workspace}/redishappy-consul-service ${name}/${configdir}/init/redishappy-consul-service.conf

    # Versioning
    echo ${version} > ${name}/${installdir}/../share/doc/redishappy-consul/VERSION
    pushd ${name}

      # rubygem: fpm
    fpm -t ${pkgtype} \
        -n ${name} \
        -v ${version}${package_version} \
        --description "${description}" \
        --url="${url}" \
        -a ${arch} \
        --category ${section} \
        --vendor ${vendor} \
        -m "${USER}@${HOSTNAME}" \
        --license "${license}" \
        --deb-upstart ../redishappy-consul-service \
        --prefix=/ \
        -s dir \
	--config-files etc/init \
	--config-files etc/redishappy-consul \
	--pre-install ../local-scripts/preinstall-redishappy-consul \
	--post-install ../local-scripts/postinstall-redishappy-consul \
	--post-uninstall ../local-scripts/postrm-redishappy-consul \
        -- .

  mv ${name}*.${pkgtype} ${origdir}/${workspace}/

  popd
}

function main() {
    makeRedisHAProxyPackage
    makeRedisConsulPackage
}

main
