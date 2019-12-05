package pui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/TNK-Studio/gortal/src/config"
	"github.com/TNK-Studio/gortal/src/core/sshd"
	"github.com/TNK-Studio/gortal/src/utils"
	"github.com/elfgzp/promptui"
	"github.com/gliderlabs/ssh"
)

// AddServer add server to config
func AddServer(sess *ssh.Session) (*string, *config.Server, error) {
	fmt.Println("Add a server.")
	namePui := promptui.Prompt{
		Label:    "Server name",
		Validate: Required("server name"),
		Stdin:    *sess,
		Stdout:   *sess,
	}

	name, err := namePui.Run()
	if err != nil {
		return nil, nil, err
	}

	hostPui := promptui.Prompt{
		Label:    "Server host",
		Validate: Required("server host"),
		Stdin:    *sess,
		Stdout:   *sess,
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
		Stdin:   *sess,
		Stdout:  *sess,
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
func AddServerSSHUser(serverKey string, sess *ssh.Session) (*string, *config.SSHUser, error) {
	fmt.Println("Add a server ssh user.")
	usernamePui := promptui.Prompt{
		Label:    "SSH username",
		Validate: Required("SSH username"),
		Stdin:    *sess,
		Stdout:   *sess,
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
		Stdin:  *sess,
		Stdout: *sess,
	}

	identityFile, err := identityFilePui.Run()
	if err != nil {
		return nil, nil, err
	}

	allowAllUserPui := promptui.Prompt{
		Label:    "Allow all user access ? yes/no",
		Validate: YesOrNo(),
		Stdin:    *sess,
		Stdout:   *sess,
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
func GetServerSSHUsersMenu(server *config.Server) func(int, *MenuItem, *ssh.Session) *[]MenuItem {
	return func(index int, menuItem *MenuItem, sess *ssh.Session) *[]MenuItem {
		menu := make([]MenuItem, 0)
		user := config.Conf.GetUserByUsername((*sess).User())
		sshUsers := (*config.Conf).GetServerSSHUsers(user, server)
		for _, sshUser := range sshUsers {
			menu = append(
				menu,
				MenuItem{
					Label: sshUser.SSHUsername,
					SelectedFunc: func(index int, menuItem *MenuItem, sess *ssh.Session) error {
						err := sshd.Connect(
							server.Host,
							server.Port,
							sshUser.SSHUsername,
							sshUser.IdentityFile,
							sess,
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
func GetServersMenu() func(int, *MenuItem, *ssh.Session) *[]MenuItem {
	return func(index int, menuItem *MenuItem, sess *ssh.Session) *[]MenuItem {
		menu := make([]MenuItem, 0)
		user := config.Conf.GetUserByUsername((*sess).User())
		servers := (*config.Conf).GetUserServers(user)
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
