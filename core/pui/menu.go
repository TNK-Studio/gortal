package pui

import (
	"errors"
	"fmt"

	"github.com/TNK-Studio/gortal/config"
	"github.com/TNK-Studio/gortal/core/sshd"
	"github.com/elfgzp/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func defaultShow(int, *MenuItem, *ssh.Session, []*MenuItem) bool { return true }

func isAdmin(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) bool {
	user := config.Conf.GetUserByUsername((*sess).User())
	if user != nil {
		return user.Admin
	}
	return false
}

func staticSubMenu(subMenu *[]*MenuItem) func(int, *MenuItem, *ssh.Session, []*MenuItem) *[]*MenuItem {
	return func(int, *MenuItem, *ssh.Session, []*MenuItem) *[]*MenuItem {
		return subMenu
	}
}

// MenuItem menu item
type MenuItem struct {
	Label             string
	Info              map[string]string
	IsShow            func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) bool
	SubMenuTitle      string
	GetSubMenu        func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) *[]*MenuItem
	SelectedFunc      func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error
	NoSubMenuInfo     string
	BackAfterSelected bool
	BackOptionLabel   string
}

// MainMenu main menu
var (
	MainMenu        *[]*MenuItem
	ListServersMenu *MenuItem
	EditUsersMenu   *MenuItem
	EditServersMenu *MenuItem
	PersonalMenu    *MenuItem

	sshUserInfoKey = "sshUserKey"
	serverInfoKey  = "serverKey"
	userInfoKey    = "userKey"
)

func init() {
	ListServersMenu = &MenuItem{
		Label:         "List servers",
		IsShow:        defaultShow,
		GetSubMenu:    GetServersMenu(),
		NoSubMenuInfo: "You don't have a server yet, please go to the 'Edit Server' menu to add a server",
	}

	EditUsersMenu = &MenuItem{
		Label:  "Edit users",
		IsShow: isAdmin,
		GetSubMenu: staticSubMenu(&[]*MenuItem{
			&MenuItem{
				Label: "Add user",
				SelectedFunc: func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
					_, _, err := CreateUser(isAdmin(index, menuItem, sess, selectedChain), false, sess)
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
			&MenuItem{
				Label: "Delete user",
				GetSubMenu: GetUsersMenu(
					func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
						userKey := fmt.Sprintf("users%d", index+1)
						user := (*config.Conf.Users)[userKey]
						if user == nil {
							return fmt.Errorf("Key '%s' of user not existed. ", userKey)
						}
						if user.Username == (*sess).User() {
							return errors.New("Can not delete current user. ")
						}
						delete(*config.Conf.Users, userKey)
						config.Conf.ReIndexUser()
						err := config.Conf.SaveTo(*config.ConfPath)
						if err != nil {
							return err
						}
						return nil
					},
				),
			},
		}),
	}

	EditServersMenu = &MenuItem{
		Label:  "Edit servers",
		IsShow: isAdmin,
		GetSubMenu: staticSubMenu(&[]*MenuItem{
			&MenuItem{
				Label: "Add server",
				SelectedFunc: func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
					var runningErr *error
					var serverKey *string

					defer func() {
						if runningErr != nil && serverKey != nil {
							server := (*config.Conf.Servers)[*serverKey]
							if server != nil {
								delete(*config.Conf.Servers, *serverKey)
							}
						}
					}()

					serverKey, _, err := AddServer(sess)
					if err != nil {
						runningErr = &err
						return err
					}
					_, _, err = AddServerSSHUser(*serverKey, sess)
					if err != nil {
						runningErr = &err
						return err
					}
					config.Conf.SaveTo(*config.ConfPath)
					if err != nil {
						runningErr = &err
						return err
					}
					return nil
				},
			},
			&MenuItem{
				Label: "Edit server",
				GetSubMenu: GetEditedServersMenu(
					staticSubMenu(&[]*MenuItem{
						&MenuItem{
							Label: "Edit server info",
							SelectedFunc: func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
								parentMenu := selectedChain[len(selectedChain)-1]
								serverKey := parentMenu.Info[serverInfoKey]
								server := (*config.Conf.Servers)[serverKey]
								newServer, err := EditServer(server, sess)
								if err != nil {
									return err
								}

								if server == nil {
									return fmt.Errorf("Key '%s' of server not existed. ", serverKey)
								}

								(*config.Conf.Servers)[serverKey] = newServer
								err = config.Conf.SaveTo(*config.ConfPath)
								if err != nil {
									return err
								}
								parentMenu.Label = newServer.Name
								return nil
							},
						},
						&MenuItem{
							Label: "Add server ssh users",
							SelectedFunc: func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
								parentMenu := selectedChain[len(selectedChain)-1]
								serverKey := parentMenu.Info[serverInfoKey]
								_, _, err := AddServerSSHUser(serverKey, sess)
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
						&MenuItem{
							Label: "Edit server ssh users",
							GetSubMenu: func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) *[]*MenuItem {
								parentMenu := selectedChain[len(selectedChain)-1]
								serverKey := parentMenu.Info[serverInfoKey]
								server := (*config.Conf.Servers)[serverKey]
								return GetEditedSSHUsersMenu(server)
							},
						},
					}),
					nil,
				),
			},
			&MenuItem{
				Label: "Delete server",
				GetSubMenu: GetEditedServersMenu(
					nil,
					func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
						serverKey := fmt.Sprintf("server%d", index+1)
						server := (*config.Conf.Servers)[serverKey]
						if server == nil {
							return fmt.Errorf("Key '%s' of server not existed. ", serverKey)
						}

						delete(*config.Conf.Servers, serverKey)
						config.Conf.ReIndexServer()
						err := config.Conf.SaveTo(*config.ConfPath)
						if err != nil {
							return err
						}
						return nil
					},
				),
			},
		}),
	}

	PersonalMenu = &MenuItem{
		Label: "Edit personal info",
		GetSubMenu: staticSubMenu(&[]*MenuItem{
			&MenuItem{
				Label: "Change your password",
				SelectedFunc: func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error {
					err := ChangePassword((*sess).User(), sess)
					if err != nil {
						return err
					}
					err = config.Conf.SaveTo(*config.ConfPath)
					if err != nil {
						return err
					}
					sshd.Info("Please login again with your new password.\n", sess)
					(*sess).Exit(0)
					sshConn := (*sess).Context().Value(ssh.ContextKeyConn).(gossh.Conn)
					sshConn.Close()
					return nil
				},
			},
		}),
	}

	MainMenu = &[]*MenuItem{
		ListServersMenu,
		EditUsersMenu,
		EditServersMenu,
		PersonalMenu,
	}
}
