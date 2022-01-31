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

func main() {
	sp := selection.New("Where do you want to install kiss?",
		selection.Choices([]string{"Drive", "Folder"}))
	sp.PageSize = 3

	choice, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}

	// do something with the final choice
	//_ = choice
	//fmt.Println(choice.Value)
	switch choice.Value {
	case "Drive":
		drive()
	case "Folder":
		fmt.Println(folder())
	}
}
func make_folder(folder string) {
	cmd := exec.Command("mkdir", "-p", folder)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}
func folder() string {
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
	return name
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
func drive() string {
	sp := selection.New("On which drive do you want to install kiss?",
		selection.Choices(GetDevices()))
	sp.PageSize = 3

	drive, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	gdisk_cmd := exec.Command("su", "--command", "gdisk /dev/"+string(drive.String))
	gdisk_cmd.Stdout = os.Stdout
	gdisk_cmd.Stderr = os.Stderr
	gdisk_cmd.Stdin = os.Stdin
	err = gdisk_cmd.Run()
	if err != nil {
		fmt.Printf("Error runing gdisk you may don't have gdisk install: %v\n", err)
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
		format_cmd := exec.Command("su", "--command", "mkfs.vfat /dev/"+string(boot_drive.String))
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
		format_cmd := exec.Command("su", "--command", "mkfs.ext4 /dev/"+string(root_drive.String))
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
	fmt.Println("mkdir -p /tmp/kiss/boot;mount /dev/" + string(root_drive.String) + " /tmp/kiss" + " /tmp/kiss;mount /dev/" + string(boot_drive.String) + " /tmp/kiss/boot" + ";chown " + user + ":" + user + " /tmp/kiss")
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
	//fmt.Println("Drive")
	return "/tmp/kiss"
}
