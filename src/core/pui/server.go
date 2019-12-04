package pui

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/TNK-Studio/gortal/src/config"
	"github.com/manifoldco/promptui"
)

// AddServer add server to config
func AddServer() (*string, *config.Server, error) {
	fmt.Println("Add a server.")
	namePui := promptui.Prompt{
		Label:    "Server name",
		Validate: Required("server name"),
	}

	name, err := namePui.Run()
	if err != nil {
		return nil, nil, err
	}

	hostPui := promptui.Prompt{
		Label:    "Server host",
		Validate: Required("server host"),
	}

	host, err := hostPui.Run()
	if err != nil {
		return nil, nil, err
	}

	portPui := promptui.Prompt{
		Label: "Server port",
		Validate: MultiValidate(
			[](func(string) error){
				Required("server port"),
				IsInt(),
			},
		),
		Default: "22",
	}

	portString, err := portPui.Run()
	if err != nil {
		return nil, nil, err
	}

	port, _ := strconv.Atoi(portString)
	key, server := config.Conf.AddServer(name, host, port)

	return &key, server, nil
}

// AddServerSSHUser add server ssh user
func AddServerSSHUser(serverKey string) (*string, *config.SSHUser, error) {
	fmt.Println("Add a server ssh user.")
	usernamePui := promptui.Prompt{
		Label:    "SSH username",
		Validate: Required("SSH username"),
	}

	username, err := usernamePui.Run()
	if err != nil {
		return nil, nil, err
	}

	identityFilePui := promptui.Prompt{
		Label: "Identity file",
		Validate: MultiValidate(
			[](func(string) error){
				Required("identity file path"),
				func(input string) error {
					input = strings.Replace(input, "~", os.Getenv("HOME"), 1)
					if !config.ConfigFileExisted(input) {
						return errors.New(fmt.Sprintf("Identity file '%s' not existed", input))
					}
					return nil
				},
			},
		),
	}

	identityFile, err := identityFilePui.Run()
	if err != nil {
		return nil, nil, err
	}

	allowAllUserPui := promptui.Prompt{
		Label:    "Allow all user access ? yes/no",
		Validate: YesOrNo(),
	}

	allAllUser, err := allowAllUserPui.Run()
	allowUsersString := ""
	if allAllUser == "no" {
		allowUsersPui := promptui.Prompt{
			Label: "Please enter all usernames separated by ','",
			Validate: MultiValidate(
				[](func(string) error){
					Required("usernames"),
					func(input string) error {
						allowUsers := strings.Split(input, ",")
						userList := make([]string, 0)
						for _, user := range *(config.Conf.Users) {
							userList = append(userList, user.Username)
						}
						for _, username := range allowUsers {
							for _, each := range userList {
								if username == each {
									break
								}
								return errors.New(fmt.Sprintf("Username '%s' of user not existed. Please choose from %v.", username, userList))
							}
						}
						return nil
					},
				},
			),
		}

		allowUsersString, err = allowUsersPui.Run()
		if err != nil {
			return nil, nil, err
		}
	}

	var allowUsers *[]string
	allowUsers = nil
	if allowUsersString != "" {
		splitedUser := strings.Split(allowUsersString, ",")
		allowUsers = &splitedUser
	}
	key, sshUser := config.Conf.AddServerSSHUser(serverKey, username, identityFile, allowUsers)
	return &key, sshUser, nil
}
