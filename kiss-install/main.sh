#!/bin/sh

set -o pipefail
log() {
    printf '\033[32m[kiss-install]->\033[m %s.\n' "$*"
}

chroot=../chroot
getroot='doas'
ver=2021.7-9
file=kiss-chroot-$ver.tar.xz
# location of the scrip
script_root=$(pwd)
chroot_script=$script_root/chroot.sh

echo $chroot
cd $chroot
chroot=$(pwd)
echo $chroot
rm -rf file
echo $url/$file
doas curl -fLO "$url/$file"
# extracting tar ball
doas tar xvf $file

cp $chroot_script .
# updating location of chroot script
chroot_script=$(pwd)/chroot.sh
chmod +x chroot.sh

log "entering chroot you now have to run ./chroot.sh"
$getroot ./bin/kiss-chroot $chroot /bin/ls
