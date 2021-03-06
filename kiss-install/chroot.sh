#!/bin/sh
set -o pipefail

profile=/root/.profile
kiss_repo_path=/root/repo/
user=root
nproc=$(nproc)
kiss=kiss

log() {
    printf '\033[32m[kiss-install-chroot]->\033[m %s.\n' "$*"
}

build() {
	. $profile
	yes '
	' | $kiss build $*	
}

runkiss() {
	su kiss -c "$*"
}

update() {
	. $profile
	yes '
	' | $kiss update
}
rebuild() {

	. $profile
	cd /var/db/kiss/installed
	yes '
	' | $kiss build *
}

# adding kiss user

#log "starting chroot script"
whoami
# clonig the kiss repo in /root/repo
git clone https://github.com/kisslinux/repo $kiss_repo_path
touch $profile
echo "export KISS_PATH='$kiss_repo_path/core:$kiss_repo_path/extra:$kiss_repo_path/wayland'
export user=root
export CFLAGS='-O3 -pipe -march=native'
export CXXFLAGS='$CFLAGS'
export MAKEFLAGS='-j$nproc'" >> $profile
build gnupg1
gpg --keyserver keyserver.ubuntu.com --recv-key 13295DAC2CF13B5C
echo trusted-key 0x13295DAC2CF13B5C >>/root/.gnupg/gpg.conf
cd $kiss_repo_path
git config merge.verifySignatures true
log "updating the system twice to make sure the update succeeds"
update; update
log "rebuilding the system (this can take a while)"
# rebuild
build perl libelf baseinit e2fsprogs dosfstools
echo '1234
1234
' | passwd root
