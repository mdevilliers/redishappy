#!/bin/bash
set -e
set -u
name=redishappy
version=${_REDISHAPPY_VERSION:-"1.0.0"}
description="RedisHappy is an automated Redis failover daemon using HaProxy and Sentinel"
url="https://github.com/mdevilliers/redishappy"
arch="all"
section="misc"
license="Apache Software License 2.0"
package_version=${_REDISHAPPY_PKGVERSION:-"-1"}
origdir="$(pwd)"
workspace="build"
pkgtype=${_PKGTYPE:-"deb"}
builddir="output"
installdir="opt"
vendor="mdevilliers"

function cleanup() {
    cd ${origdir}/${workspace}
    rm -rf ${name}*.{deb,rpm}
    rm -rf ${builddir}
}

function bootstrap() {
    cd ${origdir}/${workspace}

    # configuration directory
    mkdir -p ${builddir}/${name}/${installdir}/redishappy/config

    pushd ${builddir}
}

function build() {

    # Prepare binary at /opt/redishappy/redishappy
    cp ${origdir}/redishappy ${name}/${installdir}/redishappy/redishappy
    chmod 755 ${name}/${installdir}/redishappy/redishappy

    # Link default configuration file
    cp ${origdir}/config.json ${name}/${installdir}/redishappy/config/config.json
    cp ${origdir}/example_haproxy_template.cfg ${name}/${installdir}/redishappy/config/example_haproxy_template.cfg

    # Versioning
    echo ${version} > ${name}/${installdir}/redishappy/VERSION
    pushd ${name}
}

function mkdeb() {
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
    --deb-upstart ../../redishappy-server \
    --license "${license}" \
    --prefix=/ \
    -s dir \
    -- .
  mv ${name}*.${pkgtype} ${origdir}/${workspace}/
  popd
}

function main() {
    cleanup
    bootstrap
    build
    mkdeb
}

main