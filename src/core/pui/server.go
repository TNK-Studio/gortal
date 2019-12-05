package pui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/TNK-Studio/gortal/src/config"
	"github.com/TNK-Studio/gortal/src/core/sshd"
	"github.com/TNK-Studio/gortal/src/core/state"
	"github.com/TNK-Studio/gortal/src/utils"
	"github.com/elfgzp/promptui"
)

// AddServer add server to config
func AddServer() (*string, *config.Server, error) {
	fmt.Println("Add a server.")
	namePui := promptui.Prompt{
		Label:    "Server name",
		Validate: Required("server name"),
		Stdin:    Sess,
		Stdout:   Sess,
	}

	name, err := namePui.Run()
	if err != nil {
		return nil, nil, err
	}

	hostPui := promptui.Prompt{
		Label:    "Server host",
		Validate: Required("server host"),
		Stdin:    Sess,
		Stdout:   Sess,
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
		Stdin:   Sess,
		Stdout:  Sess,
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
		Stdin:    Sess,
		Stdout:   Sess,
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
					input = utils.FilePath(input)
					if !config.ConfigFileExisted(input) {
						return errors.New(fmt.Sprintf("Identity file '%s' not existed", input))
					}
					return nil
				},
			},
		),
		Stdin:  Sess,
		Stdout: Sess,
	}

	identityFile, err := identityFilePui.Run()
	if err != nil {
		return nil, nil, err
	}

	allowAllUserPui := promptui.Prompt{
		Label:    "Allow all user access ? yes/no",
		Validate: YesOrNo(),
		Stdin:    Sess,
		Stdout:   Sess,
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
								return errors.New(fmt.Sprintf(
									"Username '%s' of user not existed. Please choose from %v.",
									username,
									userList,
								),
								)
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

// GetServerSSHUsersMenu get server ssh users menu
func GetServerSSHUsersMenu(server *config.Server) func(int, *MenuItem) *[]MenuItem {
	return func(index int, menuItem *MenuItem) *[]MenuItem {
		menu := make([]MenuItem, 0)
		sshUsers := (*config.Conf).GetServerSSHUsers(state.CurrentUser, server)
		for _, sshUser := range sshUsers {
			menu = append(
				menu,
				MenuItem{
					Label: sshUser.SSHUsername,
					SelectedFunc: func(int, *MenuItem) error {
						err := sshd.Connect(
							server.Host,
							server.Port,
							sshUser.SSHUsername,
							sshUser.IdentityFile,
							&Sess,
						)
						if err != nil {
							return err
						}
						return nil
					},
				},
			)
		}
		return &menu
	}
}

// GetServersMenu get servers menu
func GetServersMenu() func(int, *MenuItem) *[]MenuItem {
	return func(index int, menuItem *MenuItem) *[]MenuItem {
		menu := make([]MenuItem, 0)
		servers := (*config.Conf).GetUserServers(state.CurrentUser)
		for _, server := range servers {
			menu = append(
				menu,
				MenuItem{
					Label:        server.Name,
					SubMenuTitle: fmt.Sprintf("Please select ssh user to login '%s'", server.Name),
					GetSubMenu:   GetServerSSHUsersMenu(&server),
				},
			)
		}
		return &menu
	}
}
