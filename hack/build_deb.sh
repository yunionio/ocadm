#!/usr/bin/env bash

if [ "$DEBUG" = "true" ]; then
    set -ex
    export PS4='+(${BASH_SOURCE}:${LINENO}): ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
fi
export TOPDIR="$(dirname $(dirname $(readlink -f "$BASH_SOURCE")))"
git config --global --add safe.directory $TOPDIR
export PATH=/bin:/sbin:/usr/bin:/usr/sbin
export PKG_CONFIG_PATH=/usr/lib/pkgconfig/:/usr/local/lib/pkgconfig/:/usr/local/share/pkgconfig
path=$(mktemp -d)
dest=$path/opt/yunion/bin

if [ -z "$ROOT_DIR" ]; then
	pushd $(dirname $(readlink -f "$BASH_SOURCE")) > /dev/null
	ROOT_DIR=$(cd .. && pwd)
	popd > /dev/null
fi

PACKAGE=yunion-ocadm
BUILDROOT=$path
if [ -z "$VERSION"  ]; then
    TAG=$(git describe --abbrev=0 --tags || echo 000000)
    VERSION=${TAG/\//-}
    VERSION=${VERSION/v/}
fi

RELEASE=`date +"%y%m%d%H"`
FULL_VERSION=$VERSION-$RELEASE

rm -rf $BUILDROOT
mkdir -p $BUILDROOT
mkdir -p $BUILDROOT/DEBIAN
mkdir -p $dest

case $(uname -m) in
	x86_64)
		CURRENT_ARCH=amd64
		;;
	aarch64)
		CURRENT_ARCH=arm64
		;;
esac

function build_ocadm() {
	# cd deps
	# (sh libusb_install)
	# (sh usbredir_install)
	# (sh spice_protocol_install)
	# (sh spice_install)
	# cd ..
	make -j $(grep -c ^processor /proc/cpuinfo)
    make build
    find $path
    cp -fv _output/bin/ocadm $dest
}

function build_deb() {
echo "Package: yunion-ocadm
Version: $FULL_VERSION
Section: base
Priority: optional
Architecture: $CURRENT_ARCH
Maintainer: zhangdongliang@yunionyun.com
Description: Yunion ocadm
 Yunion-ocadm $VERSION build by yunion.
" > $BUILDROOT/DEBIAN/control
chmod 0755 $BUILDROOT/DEBIAN/control
dpkg-deb --build $BUILDROOT
ls -lah $BUILDROOT.deb
mv -fv $BUILDROOT.deb yunion-ocadm-$FULL_VERSION.deb
rm -rfv $BUILDROOT
}

build_ocadm && build_deb
