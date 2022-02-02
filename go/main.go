package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
)

//const (
//	colorReset  = "\033[0m"
//	colorRed    = "\033[31m"
//	colorGreen  = "\033[32m"
//	colorYellow = "\033[33m"
//	colorBlue   = "\033[34m"
//	colorPurple = "\033[35m"
//	colorCyan   = "\033[36m"
//	colorWhite  = "\033[37m"
//)

type Install struct {
	chroot_folder string
	drive         bool
	drive_name    string
	password      string
	root_drive    string
	boot_drive    string
	getroot       string
}

func main() {
	sp := selection.New("Where do you want to install kiss?",
		selection.Choices([]string{"Install type", "Folder"}))
	sp.PageSize = 3

	choice, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}

	// do something with the final choice
	//_ = choice
	//fmt.Println(choice.Value)
	var folder string
	var kiss_install Install = Install{}
	switch choice.Value {
	case "Install type":
		kiss_install = drive()

	case "Folder":
		folder = Folder()
	}

	fmt.Println(folder)
	creat_install_script(kiss_install)
	// run /tmp/kissinstall.sh
	run_kiss_script()

}

// returns drivename bool rather if it is a drive root_drive name and boot_drive name
func drive_or_folder() bool {

	var isdrive bool
	sp := selection.New("Where do you want to install kiss?",
		selection.Choices([]string{"Install type", "Folder"}))
	sp.PageSize = 3

	choice, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
	switch choice.Value {
	case "Install type":
		isdrive = true

	case "Folder":
		isdrive = false

	}
	return isdrive
}

func run_kiss_script() {
	cmd := exec.Command("chmod", "+x", "/tmp/kissinstall.sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	cmd = exec.Command("/tmp/kissinstall.sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}
func get_install_script(install Install) string {
	if install.chroot_folder == "/" || install.chroot_folder == "" {
		fmt.Println("NOOOOOOOO")
		os.Exit(1)

	}
	create_chroot_script(install)
	return `#!/bin/sh

set -o pipefail
log() {
    printf '\033[32m[kiss-install]->\033[m %s.\n' "$*"
}

chroot="` + install.chroot_folder + `"
getroot='doas'
ver=2021.7-9
url=https://github.com/kisslinux/repo/releases/download/$ver
file=kiss-chroot-$ver.tar.xz
# location of the scrip
script_root=$(pwd)
chroot_script=/tmp/chroot.sh

echo $chroot
# now in chroot dir
cd $chroot
chroot=$(pwd)
echo $chroot
$getroot rm -rf $file
echo $url/$file
$getroot rm -rf ` + install.chroot_folder + "*" + `
$getroot curl -fLO "$url/$file"
# extracting tar ball

$getroot tar xvf $file

$getroot cp $chroot_script $chroot
# updating location of chroot script
chroot_script=$(pwd)/chroot.sh
chmod +x chroot.sh

log "entering chroot you now have to run ./chroot.sh"
$getroot ./bin/kiss-chroot $chroot `
}
func create_chroot_script(install Install) {
	var kernel string
	if install.drive {
		kernel = `git clone https://github.com/luis-07/bin-kernel /tmp/kernel
rm -rf /tmp/kernel/.git
cp /tmp/kernel/* /boot/
grub-install  --target=x86_64-efi --efi-directory=/boot --removable
grub-mkconfig -o /boot/grub/grub.cfg
echo "/dev/` + install.root_drive + ` / ext4 defaults,noatime 0 2" >> /etc/fstab`
	}
	var script string = `#!/bin/sh
set -uo pipefail

profile=/root/.profile
kiss_repo_path=/root/repo
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



update() {

	. $profile
	yes '
	' | $kiss update

}



# clonig the kiss repo in /root/repo

git clone https://github.com/kisslinux/repo $kiss_repo_path
touch $profile
echo "export KISS_PATH='$kiss_repo_path/core:$kiss_repo_path/extra:$kiss_repo_path/wayland'">> $profile
echo "export user=root" >> $profile
echo "export CFLAGS='-O0 -pipe -march=native'" >> $profile
echo "export CXXFLAGS='$CFLAGS' " >> $profile
echo "export MAKEFLAGS='-j$nproc'" >> $profile
build gnupg1
gpg --keyserver keyserver.ubuntu.com --recv-key 13295DAC2CF13B5C
echo trusted-key 0x13295DAC2CF13B5C >>/root/.gnupg/gpg.conf
cd $kiss_repo_path
git config merge.verifySignatures true
log "updating the system twice to make sure the update succeeds"
#update; update
build perl libelf baseinit e2fsprogs dosfstools grub efibootmgr
echo '` + install.password + `
` + install.password + `
' | passwd root
` + kernel
	// write the script to /tmp/chroot.sh
	err := ioutil.WriteFile("/tmp/chroot.sh", []byte(script), 0755)
	if err != nil {
		fmt.Println(err)
	}

}
func creat_install_script(install Install) {
	var install_script_location string = "/tmp/kissinstall.sh"
	err := ioutil.WriteFile(install_script_location, []byte(get_install_script(install)), 0755)
	if err != nil {
		fmt.Printf("Unable to write file: %v", err)
	}
}
func make_folder(folder string) {
	cmd := exec.Command("mkdir", "-p", folder)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}
func Folder() string {
	input := textinput.New("In which folder do you want to install kiss?")
	input.InitialValue = "/kiss"
	input.Placeholder = "Your name cannot be empty"

	name, err := input.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}

	// do something with the result
	if name[len(name)-1:] != "/" {
		name = name + "/"
	}
	make_folder(name)
	return string(name)
}
func GetDevices() []string {
	dir, err := ioutil.ReadDir("/sys/block")
	if err != nil {
		panic(err)
	}

	files := make([]string, 0)

	for _, f := range dir {
		if strings.HasPrefix(f.Name(), "sr0") || strings.HasPrefix(f.Name(), "loop") {
			continue
		}
		files = append(files, f.Name())
	}

	return files
}
func GetPartitions(device string) []string {
	dir, err := ioutil.ReadDir("/sys/block/" + device + "/")
	if err != nil {
		panic(err)
	}

	files := make([]string, 0)

	for _, f := range dir {

		if strings.HasPrefix(f.Name(), device) {
			files = append(files, f.Name())
		}

	}

	return files
}
func AreYouSure(msg string) bool {
	sp := selection.New("Are you sure you want to "+msg+"?",
		selection.Choices([]string{"Yes", "No"}))
	sp.PageSize = 3

	sure, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
	var answer bool = false
	if sure.Value == "Yes" {
		answer = true
	}
	return answer
}
func drive() Install {
	sp := selection.New("On which drive do you want to install kiss?",
		selection.Choices(GetDevices()))
	sp.PageSize = 3

	drive, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	sp = selection.New("Which tool do you want to use to partition your drive?",
		selection.Choices([]string{"gdisk", "cgdisk", "fdisk", "cfdisk", "none"}))
	sp.PageSize = 3

	disk_utility, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if disk_utility.String != "none" {
		gdisk_cmd := exec.Command("su", "--command", disk_utility.String+" /dev/"+string(drive.String))
		gdisk_cmd.Stdout = os.Stdout
		gdisk_cmd.Stderr = os.Stderr
		gdisk_cmd.Stdin = os.Stdin
		err = gdisk_cmd.Run()
		if err != nil {
			fmt.Printf("Error runing gdisk you may don't have gdisk install: %v\n", err)
		}
	}
	sp = selection.New("Select boot partition",
		selection.Choices(GetPartitions(drive.String)))
	sp.PageSize = 3

	boot_drive, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
	if AreYouSure("you want to format " + boot_drive.String + " to vfat") {
		format_cmd := exec.Command("su", "--command", "umount /dev/"+boot_drive.String+";mkfs.vfat /dev/"+string(boot_drive.String))
		format_cmd.Stdout = os.Stdout
		format_cmd.Stderr = os.Stderr
		format_cmd.Stdin = os.Stdin
		err = format_cmd.Run()
		if err != nil {
			fmt.Printf("Error formatting boot drive: %v\n", err)
		}
	} else {
		fmt.Println("Not formatting")
	}
	sp = selection.New("Select root partition",
		selection.Choices(GetPartitions(drive.String)))
	sp.PageSize = 3

	root_drive, err := sp.RunPrompt()

	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
	if AreYouSure("you want to format " + root_drive.String + " to ext4") {
		format_cmd := exec.Command("su", "--command", "umount /dev/"+root_drive.String+";mkfs.ext4 /dev/"+string(root_drive.String))
		format_cmd.Stdout = os.Stdout
		format_cmd.Stderr = os.Stderr
		format_cmd.Stdin = os.Stdin
		err = format_cmd.Run()
		if err != nil {
			fmt.Printf("Error formatting root drive: %v\n", err)
		}
	} else {
		fmt.Println("Not formatting")
	}
	user := os.Getenv("USER")
	mount_cmd := exec.Command("su", "--command", "mkdir -p /tmp/kiss/boot;umount /tmp/kiss/boot;umount /tmp/kiss;mount /dev/"+string(root_drive.String)+" /tmp/kiss;mount /dev/"+string(boot_drive.String)+" /tmp/kiss/boot"+";chown "+user+":"+user+" /tmp/kiss")
	mount_cmd.Stdout = os.Stdout
	mount_cmd.Stderr = os.Stderr
	mount_cmd.Stdin = os.Stdin
	err = mount_cmd.Run()
	if err != nil {
		fmt.Printf("Error mount boot or root drive: %v\n", err)
	}
	fmt.Println(root_drive.String)
	fmt.Println(boot_drive.String)
	//fmt.Println(drive.Value)
	//fmt.Println("Install type")

	return Install{
		chroot_folder: "/tmp/kiss",
		drive:         true,
		drive_name:    drive.String,
		password:      ask_for_password(),
		root_drive:    root_drive.String,
		boot_drive:    boot_drive.String,
	}
}

func ask_for_password() string {
	input := textinput.New("Enter the root password:")
	input.Placeholder = "minimum 4 characters"
	input.Validate = func(s string) bool { return len(s) >= 4 }
	input.Hidden = true

	password, err := input.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
	input = textinput.New("Confirm the root password:")
	input.Placeholder = "minimum 4 characters"
	input.Validate = func(s string) bool { return len(s) >= 4 }
	input.Validate = func(s string) bool { return password == s }
	input.Hidden = true

	_, err = input.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
	// do something with the result

	return password
}
