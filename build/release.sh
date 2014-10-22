#!/bin/bash
set -e
set -u

version=${_REDISHAPPY_VERSION:-"1.0.0"}
url="https://github.com/mdevilliers/redishappy"
arch="all"
section="misc"
license="Apache Software License 2.0"
package_version=${_REDISHAPPY_PKGVERSION:-"1"}
origdir="$(pwd)"
workspace="build"
pkgtype=${_PKGTYPE:-"deb"}
builddir="output"
installdir="opt"
configdir="var"
vendor="mdevilliers"

function makeRedisHAProxyPackage() {

    name=redishappy-haproxy
    description="RedisHappy HAProxy is an automated Redis failover daemon integrating Redis Sentinel with HAProxy"

    cd ${origdir}/${workspace}
    rm -rf ${name}*.{deb,rpm}
    rm -rf ${builddir}

    mkdir -p ${name}/${installdir}/redishappy
    mkdir -p ${name}/${configdir}/redishappy

    cp ${origdir}/redis-haproxy ${name}/${installdir}/redishappy/redis-haproxy
    chmod 755 ${name}/${installdir}/redishappy/redis-haproxy

    cp ${origdir}/main/redis-haproxy/config.json ${name}/${configdir}/redishappy/config.json
    cp ${origdir}/main/redis-haproxy/example_haproxy_template.cfg ${name}/${configdir}/redishappy/example_haproxy_template.cfg

    # Versioning
    echo ${version} > ${name}/${installdir}/redishappy/VERSION
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
        --deb-upstart ../../redis-haproxy-server \
        --prefix=/ \
        -s dir \
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

    mkdir -p ${name}/${installdir}/redishappy
    mkdir -p ${name}/${configdir}/redishappy

    cp ${origdir}/redis-consul ${name}/${installdir}/redishappy/redis-consul
    chmod 755 ${name}/${installdir}/redishappy/redis-consul

    cp ${origdir}/main/redis-consul/config.json ${name}/${configdir}/redishappy/config.json

    # Versioning
    echo ${version} > ${name}/${installdir}/redishappy/VERSION
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
        --deb-upstart ../../redis-consul-server \
        --prefix=/ \
        -s dir \
        -- .

  mv ${name}*.${pkgtype} ${origdir}/${workspace}/

  popd
}


function main() {
    makeRedisHAProxyPackage
    makeRedisConsulPackage
}

main