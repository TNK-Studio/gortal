package pui

import (
	"errors"
	"fmt"

	"github.com/TNK-Studio/gortal/src/config"
	"github.com/gliderlabs/ssh"
)

// MainMenu main menu
var MainMenu *[]MenuItem

func defaultShow(int, *MenuItem, *ssh.Session) bool { return true }

func isAdmin(index int, menuItem *MenuItem, sess *ssh.Session) bool {
	user := config.Conf.GetUserByUsername((*sess).User())
	if user != nil {
		return user.Admin
	}
	return false
}

func staticSubMenu(subMenu *[]MenuItem) func(int, *MenuItem, *ssh.Session) *[]MenuItem {
	return func(int, *MenuItem, *ssh.Session) *[]MenuItem {
		return subMenu
	}
}

// MenuItem menu item
type MenuItem struct {
	Label             string
	IsShow            func(int, *MenuItem, *ssh.Session) bool
	SubMenuTitle      string
	GetSubMenu        func(int, *MenuItem, *ssh.Session) *[]MenuItem
	SelectedFunc      func(int, *MenuItem, *ssh.Session) error
	BackAfterSelected bool
	BackOptionLabel   string
}

func init() {
	MainMenu = &[]MenuItem{
		MenuItem{
			Label:      "List servers",
			IsShow:     defaultShow,
			GetSubMenu: GetServersMenu(),
		},
		MenuItem{
			Label:  "Edit users",
			IsShow: isAdmin,
			GetSubMenu: staticSubMenu(&[]MenuItem{
				MenuItem{
					Label: "Add user",
					SelectedFunc: func(index int, menuItem *MenuItem, sess *ssh.Session) error {
						_, _, err := CreateUser(isAdmin(index, menuItem, sess), false, sess)
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
				MenuItem{
					Label: "Delete user",
					GetSubMenu: GetUsersMenu(
						func(index int, menuItem *MenuItem, sess *ssh.Session) error {
							userKey := fmt.Sprintf("user%d", index+1)
							user := (*config.Conf.Users)[userKey]
							if user == nil {
								return errors.New(fmt.Sprintf("Key '%s' of user not existed. ", userKey))
							}
							if user.Username == (*sess).User() {
								return errors.New("Can not delete current user. ")
							}
							delete(*config.Conf.Users, userKey)
							config.Conf.ReIndexUser()
							config.Conf.SaveTo(*config.ConfPath)
							return nil
						},
					),
				},
			}),
		},
		MenuItem{
			Label:  "Edit servers info",
			IsShow: isAdmin,
			GetSubMenu: staticSubMenu(&[]MenuItem{
				MenuItem{
					Label: "Add server",
					SelectedFunc: func(index int, menuItem *MenuItem, sess *ssh.Session) error {
						serverKey, _, err := AddServer(sess)
						if err != nil {
							return err
						}
						_, _, err = AddServerSSHUser(*serverKey, sess)
						if err != nil {
							return err
						}
						config.Conf.SaveTo(*config.ConfPath)
						if err != nil {
							return err
						}
						return nil
					},
				},
				MenuItem{
					Label: "Delete server",
					GetSubMenu: GetEditedServersMenu(
						func(index int, menuItem *MenuItem, sess *ssh.Session) error {
							serverKey := fmt.Sprintf("server%d", index+1)
							server := (*config.Conf.Servers)[serverKey]
							if server == nil {
								return errors.New(fmt.Sprintf("Key '%s' of server not existed. ", serverKey))
							}

							delete(*config.Conf.Servers, serverKey)
							config.Conf.ReIndexServer()
							config.Conf.SaveTo(*config.ConfPath)
							return nil
						},
					),
				},
			}),
		},
	}
}
