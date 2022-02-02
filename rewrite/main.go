package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
)

// TODO: error handling

const (
	colorBlue  = "\033[34m"
	colorWhite = "\033[37m"
)

type Install struct {
	Drive              bool
	Install_Location   string
	Install_Drive      string
	Password           string
	last_step          []string
	last_was_next_menu bool
	boot_drive         string
	root_drive         string
	hostname           string
	users              []User
}

type User struct {
	Username string
	Password string
}

func main() {
	install := new(Install)

	*install = drive_or_folder(*install)
	if !install.last_was_next_menu {
		*install = nextMenu(*install)
	}

	*install = select_drive(*install)
	if !install.last_was_next_menu {
		*install = nextMenu(*install)
	}

	*install = select_boot_partition(*install)
	if !install.last_was_next_menu {
		*install = nextMenu(*install)
	}

	*install = select_root_partition(*install)
	if !install.last_was_next_menu {
		*install = nextMenu(*install)
	}

	*install = ask_for_hostname(*install)
	if !install.last_was_next_menu {
		*install = nextMenu(*install)
	}

	*install = ask_for_password(*install)
	if !install.last_was_next_menu {
		*install = nextMenu(*install)
	}
}

// retunrs the last element of the slice and rmoves it
func last_element_slice_string_and_remove(slice []string) (string, []string) {
	var last string = slice[len(slice)-1]
	return last, slice[:len(slice)-1]
}

func visualize_config(install Install) {
	// print a nice looking table
	if install.Drive {
		fmt.Println("Install type:", "Install type")
	} else {
		fmt.Println("Install type:", "Folder")
	}
}

// executes function based on a strings value
func string_to_func(install Install, String string) Install {
	// add here new options
	switch String {
	case "Install type":
		install = drive_or_folder(install)
	case "Password":
		install = ask_for_password(install)
	case "Select drive":
		install = select_drive(install)
	}
	return install
}

func select_config_menu(install Install) Install {
	install.last_was_next_menu = false
	sp := selection.New("What do you want to configure?",
		// add here new options
		selection.Choices([]string{"Install type", "Password", "Select drive"}))
	sp.PageSize = 3

	choice, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
	install = string_to_func(install, choice.String)

	install = nextMenu(install)
	return install
}

func back(install Install) Install {
	install.last_was_next_menu = false
	last := ""
	last, install.last_step = last_element_slice_string_and_remove(install.last_step)

	install = string_to_func(install, last)
	install = nextMenu(install)
	return install
}

func configure(install Install) Install {
	//visualize_config(install)
	select_config_menu(install)

	install = nextMenu(install)
	return install

}

//
// beginnig of config menus
//

func ask_for_password(install Install) Install {

	input := textinput.New("Enter the root password for the installation:")
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
	install.Password = password
	install.last_was_next_menu = false
	install.last_step = append(install.last_step, "Password")
	return install
}

func drive_or_folder(install Install) Install {
	sp := selection.New("Where do you want to install kiss?",
		selection.Choices([]string{"Drive", "Folder"}))
	sp.PageSize = 3

	choice, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}

	if choice.String == "Drive" {
		install.Drive = true
	} else {
		install.Drive = false
	}
	install.last_step = append(install.last_step, "Install type")
	install.last_was_next_menu = false
	return install
}

func select_drive(install Install) Install {
	install.last_was_next_menu = false
	sp := selection.New("Select the drive you want to install kiss on:",
		selection.Choices(list_drives()))
	sp.PageSize = 3

	drive, err := sp.RunPrompt()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	install.Install_Drive = drive.String
	install.last_step = append(install.last_step, "Select drive")
	install.last_was_next_menu = false
	return install
}

func select_boot_partition(install Install) Install {
	sp := selection.New("Select boot partition",
		selection.Choices(list_partitions(install.Install_Drive)))
	sp.PageSize = 3

	boot_drive, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
	install.last_was_next_menu = false
	install.boot_drive = boot_drive.String
	return install
}

func select_root_partition(install Install) Install {
	sp := selection.New("Select root partition",
		selection.Choices(list_partitions(install.Install_Drive)))
	sp.PageSize = 3

	root_drive, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
	install.last_was_next_menu = false
	install.root_drive = root_drive.String
	return install
}

func ask_for_hostname(install Install) Install {
	input := textinput.New("Enter a hostname:")
	input.InitialValue = "kiss"
	input.Placeholder = "Hostname cannot be empty"

	hostname, err := input.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}

	// do something with the result
	install.hostname = hostname
	install.last_was_next_menu = false
	return install
}

//
// end of config menus
//

// user menu

func create_new_user(install Install) Install {
	var user User
	install, user = ask_for_username(user, install)
	install.last_was_next_menu = false
	return install
}

func ask_for_username(user User, install Install) (Install, User) {
	input := textinput.New("What is your username?")
	input.InitialValue = "glenda"
	input.Placeholder = "Your username cannot be empty"
	username, err := input.RunPrompt()
	if err != nil {
		fmt.Printf("Error: (user creation) %v\n", err)
	}
	user.Username = username
	install.last_was_next_menu = false
	return install, user
}

// partitioning

func list_partitions(device string) []string {
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

func list_drives() []string {
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

// end partitioning

func nextMenu(install Install) Install {
	// if install.last_step is empty
	if install.last_step == nil || len(install.last_step) == 0 {
		fmt.Println("Error no last step, please report this bug to the maintainer")
		os.Exit(1)
	}
	sp := selection.New("Next step or back ("+colorBlue+install.last_step[len(install.last_step)-1]+colorWhite+") ?",
		selection.Choices([]string{"Next", "Back", "Configure"}))

	sp.PageSize = 3

	answer, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
	switch answer.String {
	case "Next":
		install.last_was_next_menu = true
	case "Back":
		back(install)
	case "Configure":
		configure(install)
	}
	return install
}
