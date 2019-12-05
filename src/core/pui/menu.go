package pui

import (
	"github.com/TNK-Studio/gortal/src/config"
	"github.com/gliderlabs/ssh"
)

// MainMenu main menu
var MainMenu *[]MenuItem

func defaultShow(int, *MenuItem, *ssh.Session) bool { return true }

func isAdmin(index int, menuItem *MenuItem, sess *ssh.Session) bool {
	user := config.Conf.GetUserByUsername((*sess).User())
	return user.Admin
}

func staticSubMenu(subMenu *[]MenuItem) func(int, *MenuItem, *ssh.Session) *[]MenuItem {
	return func(int, *MenuItem, *ssh.Session) *[]MenuItem {
		return subMenu
	}
}

// MenuItem menu item
type MenuItem struct {
	Label           string
	IsShow          func(int, *MenuItem, *ssh.Session) bool
	SubMenuTitle    string
	GetSubMenu      func(int, *MenuItem, *ssh.Session) *[]MenuItem
	SelectedFunc    func(int, *MenuItem, *ssh.Session) error
	backOptionLabel string
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
				MenuItem{Label: "Edit user info"},
				MenuItem{Label: "Delete user"},
			}),
		},
		MenuItem{
			Label:  "Edit servers info",
			IsShow: isAdmin,
			GetSubMenu: staticSubMenu(&[]MenuItem{
				MenuItem{Label: "Add server"},
				MenuItem{Label: "Edit server info"},
				MenuItem{Label: "Delete server"},
			}),
		},
	}
}
