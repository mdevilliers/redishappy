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
vendor="mdevilliers"

function buildPackage() {

    name=$1
    description=$2

    cd ${origdir}/${workspace}
    rm -rf ${name}*.{deb,rpm}
    rm -rf ${builddir}

    mkdir -p ${name}/${installdir}/redishappy

    cp ${origdir}/redis-consul ${name}/${installdir}/redishappy/redis-consul
    chmod 755 ${name}/${installdir}/redishappy/redis-consul

    cp ${origdir}/main/redis-consul/config.json ${name}/${installdir}/redishappy/config.json

    # Versioning
    echo ${version} > ${name}/${installdir}/redishappy/VERSION
    pushd ${name}

      # rubygem: fpm
  #  --deb-upstart ../../redishappy-server \
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
        --prefix=/ \
        -s dir \
        -- .

  mv ${name}*.${pkgtype} ${origdir}/${workspace}/

  popd
}

function main() {

    buildPackage "redishappy-haproxy" "RedisHappy HAProxy is an automated Redis failover daemon integrating Redis Sentinel with HAProxy"
    buildPackage "redishappy-consul" "RedisHappy Consul is an automated Redis failover daemon integrating Redis Sentinel with Consul"

}

main