package pui

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/TNK-Studio/gortal/config"
	"github.com/TNK-Studio/gortal/core/sshd"
	"github.com/TNK-Studio/gortal/utils"
	"github.com/TNK-Studio/gortal/utils/logger"
	"github.com/elfgzp/promptui"
	"github.com/gliderlabs/ssh"
)

// AddServer add server to config
func AddServer(sess *ssh.Session) (*string, *config.Server, error) {
	logger.Logger.Info("Add a server.")
	stdio := utils.SessIO(sess)
	namePui := serverNamePrompt("", stdio)

	name, err := namePui.Run()
	if err != nil {
		return nil, nil, err
	}

	hostPui := serverHostPrompt("", stdio)

	host, err := hostPui.Run()
	if err != nil {
		return nil, nil, err
	}

	portPui := serverPortPrompt("22", stdio)

	portString, err := portPui.Run()
	if err != nil {
		return nil, nil, err
	}

	port, _ := strconv.Atoi(portString)
	key, server := config.Conf.AddServer(name, host, port)

	return &key, server, nil
}

// EditServer EditServer
func EditServer(server *config.Server, sess *ssh.Session) (*config.Server, error) {
	stdio := utils.SessIO(sess)
	namePui := serverNamePrompt(server.Name, stdio)

	name, err := namePui.Run()
	if err != nil {
		return nil, err
	}

	hostPui := serverHostPrompt(server.Host, stdio)

	host, err := hostPui.Run()
	if err != nil {
		return nil, err
	}

	portPui := serverPortPrompt(fmt.Sprintf("%d", server.Port), stdio)

	portString, err := portPui.Run()
	if err != nil {
		return nil, err
	}

	port, _ := strconv.Atoi(portString)

	newServer := &config.Server{
		Name:     name,
		Host:     host,
		Port:     port,
		SSHUsers: (*server).SSHUsers,
	}
	return newServer, nil
}

// AddServerSSHUser add server ssh user
func AddServerSSHUser(serverKey string, sess *ssh.Session) (*string, *config.SSHUser, error) {
	logger.Logger.Info("Add a server ssh user.")
	var runningErr *error

	stdio := utils.SessIO(sess)
	usernamePui := sshUserNamePrompt("", stdio)
	server := (*config.Conf.Servers)[serverKey]
	if server == nil {
		return nil, nil, fmt.Errorf("Server of server key '%s' not existed. ", serverKey)
	}

	username, err := usernamePui.Run()
	if err != nil {
		return nil, nil, err
	}

	hasIdentityFilePui := promptui.Prompt{
		Label:    "Has identity file ? yes/no",
		Validate: YesOrNo(),
		Stdin:    stdio,
		Stdout:   stdio,
	}

	hasIdentityFile, err := hasIdentityFilePui.Run()
	if err != nil {
		return nil, nil, err
	}

	var identityFile string
	var identityFilePui promptui.Prompt
	if hasIdentityFile == "yes" {
		identityFilePui = identityFilePrompt(
			"Enter your identity file path",
			"",
			stdio,
			FileExited("Identity File"),
		)

		identityFile, err = identityFilePui.Run()
		if err != nil {
			return nil, nil, err
		}
	} else {
		defaultPath := fmt.Sprintf("~/.ssh/id_%s", strings.ReplaceAll(server.Host, ".", "_"))
		identityFilePui = identityFilePrompt(
			"Enter new identity file path",
			defaultPath,
			stdio,
			FileNotExited("Identity File"),
			IsNotDir(),
			func(input string) error {
				var file *os.File
				fileName := utils.FilePath(input)

				defer func() {
					if file != nil {
						file.Close()
					}

					if utils.FileExited(fileName) {
						err = os.Remove(fileName)
						if err != nil {
							logger.Logger.Error(err)
						}
					}
				}()

				file, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
				if err != nil {
					return err
				}
				return nil
			},
		)
		identityFile, err = identityFilePui.Run()
		if err != nil {
			return nil, nil, err
		}

		runningErr = &err
		defer func() {
			if *runningErr != nil && identityFile != "" {
				if utils.FileExited(identityFile) {
					os.Remove(utils.FilePath(identityFile))
				}
			}
		}()

		_, pubKeyFile, err := sshd.GenKey(identityFile)
		if err != nil {
			return nil, nil, err
		}

		times := 0
		for {
			if times > 2 {
				err = errors.New("Failed to copy public key to the server. ")
				runningErr = &err
				return nil, nil, err
			}
			serverPasswdPui := promptui.Prompt{
				Label:  "Server password (use to send your public key to the server). ",
				Mask:   '*',
				Stdin:  stdio,
				Stdout: stdio,
			}

			serverPasswd, err := serverPasswdPui.Run()
			if err != nil {
				sshd.ErrorInfo(err, sess)
				times++
				continue
			}

			_, err = sshd.CopyID(
				username,
				server.Host,
				server.Port,
				serverPasswd,
				pubKeyFile,
			)
			if err != nil {
				sshd.ErrorInfo(err, sess)
				times++
				continue
			}

			break
		}
	}

	allowAllUserPui := allowAllUserPrompt("", stdio)

	allAllUser, err := allowAllUserPui.Run()
	if err != nil {
		runningErr = &err
		return nil, nil, err
	}

	allowUsersStr := ""
	if allAllUser == "no" {
		allowUsersPui := allowUsersPrompt("", stdio)

		allowUsersStr, err = allowUsersPui.Run()
		if err != nil {
			runningErr = &err
			return nil, nil, err
		}
	}

	var allowUsers *[]string
	allowUsers = nil
	if allowUsersStr != "" {
		splitedUser := strings.Split(allowUsersStr, ",")
		allowUsers = &splitedUser
	}
	key, sshUser := config.Conf.AddServerSSHUser(serverKey, username, identityFile, allowUsers)
	return &key, sshUser, nil
}

// GetServerSSHUsersMenu get server ssh users menu
func GetServerSSHUsersMenu(server *config.Server) func(int, *MenuItem, *ssh.Session, []*MenuItem) *[]*MenuItem {
	return func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) *[]*MenuItem {
		menu := make([]*MenuItem, 0)
		user := config.Conf.GetUserByUsername((*sess).User())
		sshUsers := (*config.Conf).GetServerSSHUsers(user, server)
		for sshUserKey, sshUser := range sshUsers {
			info := make(map[string]string)
			info[sshUserInfoKey] = sshUserKey
			menu = append(
				menu,
				&MenuItem{
					Label: sshUser.SSHUsername,
					Info:  info,
					SelectedFunc: func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
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
func GetServersMenu() func(int, *MenuItem, *ssh.Session, []*MenuItem) *[]*MenuItem {
	return func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) *[]*MenuItem {
		menu := make([]*MenuItem, 0)
		user := config.Conf.GetUserByUsername((*sess).User())
		servers := (*config.Conf).GetUserServers(user)
		for serverKey, server := range servers {
			info := make(map[string]string, 0)
			info[serverInfoKey] = serverKey
			menu = append(
				menu,
				&MenuItem{
					Label:        fmt.Sprintf("%s: %s", serverKey, server.Name),
					Info:         info,
					SubMenuTitle: fmt.Sprintf("Please select ssh user to login '%s'", server.Name),
					GetSubMenu:   GetServerSSHUsersMenu(server),
				},
			)
		}
		return &menu
	}
}

// GetEditedServersMenu get servers menu
func GetEditedServersMenu(
	getSubMenu func(int, *MenuItem, *ssh.Session, []*MenuItem) *[]*MenuItem,
	selectedFunc func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error,
) func(int, *MenuItem, *ssh.Session, []*MenuItem) *[]*MenuItem {
	return func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) *[]*MenuItem {
		menu := make([]*MenuItem, 0)
		serverKeys := make([]string, 0)
		for serverKey := range *config.Conf.Servers {
			serverKeys = append(serverKeys, serverKey)
		}
		sort.Strings(serverKeys)
		if len(serverKeys) < 1 {
			return &menu
		}
		for _, serverKey := range serverKeys {
			server := (*config.Conf.Servers)[serverKey]
			info := make(map[string]string)
			info[serverInfoKey] = serverKey

			menu = append(
				menu,
				&MenuItem{
					Label:             fmt.Sprintf("%s: %s", serverKey, server.Name),
					Info:              info,
					SubMenuTitle:      fmt.Sprintf("Please select. "),
					GetSubMenu:        getSubMenu,
					SelectedFunc:      selectedFunc,
					BackAfterSelected: true,
				},
			)
		}
		return &menu
	}
}

// EditSSHUser EditSSHUser
func EditSSHUser(server *config.Server, sshUser *config.SSHUser, sess *ssh.Session) (*config.SSHUser, error) {
	logger.Logger.Info("Delete ssh user")
	var runningErr *error

	stdio := utils.SessIO(sess)
	usernamePui := sshUserNamePrompt(sshUser.SSHUsername, stdio)

	username, err := usernamePui.Run()
	if err != nil {
		return nil, err
	}

	hasIdentityFilePui := promptui.Prompt{
		Label:    "Has identity file ? yes/no",
		Validate: YesOrNo(),
		Default:  utils.If(utils.FileExited(sshUser.IdentityFile), "yes", "no").(string),
		Stdin:    stdio,
		Stdout:   stdio,
	}

	hasIdentityFile, err := hasIdentityFilePui.Run()
	if err != nil {
		return nil, err
	}

	var identityFile string
	var identityFilePui promptui.Prompt
	if hasIdentityFile == "yes" {
		identityFilePui = identityFilePrompt(
			"Enter your identity file path",
			sshUser.IdentityFile,
			stdio,
			FileExited("Identity File"),
		)

		identityFile, err = identityFilePui.Run()
		if err != nil {
			return nil, err
		}
	} else {
		defaultPath := fmt.Sprintf("~/.ssh/id_%s", strings.ReplaceAll(server.Host, ".", "_"))
		identityFilePui = identityFilePrompt(
			"Enter new identity file path",
			defaultPath,
			stdio,
			FileNotExited("Identity File"),
			IsNotDir(),
			func(input string) error {
				var file *os.File
				fileName := utils.FilePath(input)

				defer func() {
					if file != nil {
						file.Close()
					}

					if utils.FileExited(fileName) {
						err = os.Remove(fileName)
						if err != nil {
							logger.Logger.Error(err)
						}
					}
				}()

				file, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
				if err != nil {
					return err
				}
				return nil
			},
		)
		identityFile, err = identityFilePui.Run()
		if err != nil {
			return nil, err
		}

		runningErr = &err
		defer func() {
			if *runningErr != nil && identityFile != "" {
				if utils.FileExited(identityFile) {
					os.Remove(utils.FilePath(identityFile))
				}
			}
		}()

		_, pubKeyFile, err := sshd.GenKey(identityFile)
		if err != nil {
			return nil, err
		}

		times := 0
		for {
			if times > 2 {
				err = errors.New("Failed to copy public key to the server. ")
				runningErr = &err
				return nil, err
			}
			serverPasswdPui := promptui.Prompt{
				Label:  "Server password (use to send your public key to the server). ",
				Mask:   '*',
				Stdin:  stdio,
				Stdout: stdio,
			}

			serverPasswd, err := serverPasswdPui.Run()
			if err != nil {
				sshd.ErrorInfo(err, sess)
				times++
				continue
			}

			_, err = sshd.CopyID(
				username,
				server.Host,
				server.Port,
				serverPasswd,
				pubKeyFile,
			)
			if err != nil {
				sshd.ErrorInfo(err, sess)
				times++
				continue
			}

			break
		}
	}

	allowAllUserPui := allowAllUserPrompt(
		utils.If(sshUser.AllowUsers == nil || len(*sshUser.AllowUsers) <= 0, "yes", "no").(string),
		stdio,
	)

	allAllUser, err := allowAllUserPui.Run()
	if err != nil {
		runningErr = &err
		return nil, err
	}

	allowUsersStr := ""
	if allAllUser == "no" {
		defaultUsers := ""
		if sshUser.AllowUsers != nil {
			defaultUsers = strings.Join(*sshUser.AllowUsers, ",")
		}
		allowUsersPui := allowUsersPrompt(defaultUsers, stdio)
		allowUsersStr, err = allowUsersPui.Run()
		if err != nil {
			runningErr = &err
			return nil, err
		}
	}

	var allowUsers *[]string
	allowUsers = nil
	if allowUsersStr != "" {
		splitedUser := strings.Split(allowUsersStr, ",")
		allowUsers = &splitedUser
	}

	newSSHUser := &config.SSHUser{
		SSHUsername:  username,
		IdentityFile: identityFile,
		AllowUsers:   allowUsers,
	}

	return newSSHUser, err
}

// DelSSHUser DelSSHUser
func DelSSHUser(server *config.Server, sshUserKey string, sess *ssh.Session) error {
	if (len(*server.SSHUsers)) <= 1 {
		return errors.New("The server requires at least one ssh user. ")
	}
	if sshUser := (*server.SSHUsers)[sshUserKey]; sshUser != nil {
		delete(*server.SSHUsers, sshUserKey)
	} else {
		return fmt.Errorf("SSH user of key '%s' does not exited. ", sshUserKey)
	}
	return nil
}

// GetEditedSSHUsersMenu GetEditedSSHUsersMenu
func GetEditedSSHUsersMenu(server *config.Server) *[]*MenuItem {
	menu := make([]*MenuItem, 0)

	for sshUserKey, sshUser := range *server.SSHUsers {
		info := make(map[string]string)
		info[sshUserInfoKey] = sshUserKey
		menu = append(
			menu,
			&MenuItem{
				Label: sshUser.SSHUsername,
				Info:  info,
				GetSubMenu: staticSubMenu(&[]*MenuItem{
					&MenuItem{
						Label: "Edit ssh user",
						SelectedFunc: func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
							parent := selectedChain[len(selectedChain)-1]
							sshUserKey := parent.Info[sshUserInfoKey]
							sshUser := (*server.SSHUsers)[sshUserKey]
							newSSHUser, err := EditSSHUser(server, sshUser, sess)
							if err != nil {
								return err
							}
							(*server.SSHUsers)[sshUserKey] = newSSHUser
							err = config.Conf.SaveTo(*config.ConfPath)
							if err != nil {
								return err
							}
							return nil
						},
					},
					&MenuItem{
						Label:             "Delete ssh user",
						BackAfterSelected: true,
						SelectedFunc: func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
							parent := selectedChain[len(selectedChain)-1]
							sshUserKey := parent.Info[sshUserInfoKey]
							err := DelSSHUser(server, sshUserKey, sess)
							if err != nil {
								return err
							}
							err = config.Conf.SaveTo(*config.ConfPath)
							if err != nil {
								return err
							}
							return nil
						},
					},
				}),
			},
		)
	}
	return &menu
}
