package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/TNK-Studio/gortal/src/config"
	"github.com/TNK-Studio/gortal/src/core/pui"
	"github.com/TNK-Studio/gortal/src/core/state"
)

var (
	username *string
)

func init() {
	config.ConfPath = flag.String("c", fmt.Sprintf("%s%s", os.Getenv("HOME"), "/.gortal.yml"), "Config file")
	username = flag.String("u", "", "Username")
}

func setupConfig() {
	fmt.Println("Config file not found. Setup config.", *config.ConfPath)
	_, user, err := pui.CreateUser(false, true)
	if err != nil {
		return
	}
	serverKey, _, err := pui.AddServer()
	if err != nil {
		return
	}
	_, _, err = pui.AddServerSSHUser(*serverKey)
	if err != nil {
		return
	}
	config.Conf.SaveTo(*config.ConfPath)
	state.CurrentUser = user
}

func setCurrentUser() {
	if state.CurrentUser == nil {
		if *username == "" {
			fmt.Printf("Please specify a username")
			return
		}

		for _, user := range *config.Conf.Users {
			if user.Username == *username {
				state.CurrentUser = user
				break
			}
		}
		if state.CurrentUser == nil {
			fmt.Printf("Usename '%s' not existed\n", *username)
			return
		}
	}

	fmt.Printf("Current user is '%s'\n", state.CurrentUser.Username)
}

func main() {
	flag.Parse()

	if *config.ConfPath == "" {
		fmt.Println("Please specify a config file.")
		return
	}
	fmt.Println("Read config file", *config.ConfPath)
	if !config.ConfigFileExisted(*config.ConfPath) {
		setupConfig()
	} else {
		config.Conf.ReadFrom(*config.ConfPath)
		if config.Conf.Users == nil || len(*config.Conf.Users) < 1 {
			_, user, err := pui.CreateUser(false, true)
			if err != nil {
				return
			}
			state.CurrentUser = user
		}
	}
	setCurrentUser()
	pui.ShowMainMenu()
}
