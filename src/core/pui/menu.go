package pui

import (
	"fmt"

	"github.com/TNK-Studio/gortal/src/config"
	"github.com/TNK-Studio/gortal/src/core/state"
	"github.com/TNK-Studio/gortal/src/utils/logger"
	"github.com/manifoldco/promptui"
)

// MainMenu main menu
var MainMenu *[]MenuItem

func defaultShow(*MenuItem) bool { return true }

func isAdmin(*MenuItem) bool { return state.CurrentUser.Admin }

func staticSubMenu(subMenu *[]MenuItem) func(*MenuItem) *[]MenuItem {
	return func(*MenuItem) *[]MenuItem {
		return subMenu
	}
}

func getListserversMenu() func(*MenuItem) *[]MenuItem {
	return func(*MenuItem) *[]MenuItem {
		menu := make([]MenuItem, 0)
		for _, server := range *config.Conf.Servers {
			show := false
		loop:
			for _, sshUser := range *server.SSHUsers {
				if sshUser.AllowUsers == nil {
					show = true
					break loop
				}

				for _, username := range *sshUser.AllowUsers {
					if state.CurrentUser.Username == username {
						show = true
						break loop
					}
				}
			}
			menu = append(
				menu,
				MenuItem{
					Label:  server.Name,
					IsShow: func(*MenuItem) bool { return show },
				},
			)
		}
		return &menu
	}
}

func init() {
	MainMenu = &[]MenuItem{
		MenuItem{
			Label:      "List servers",
			IsShow:     defaultShow,
			GetSubMenu: getListserversMenu(),
		},
		MenuItem{
			Label:  "Edit users",
			IsShow: isAdmin,
			GetSubMenu: staticSubMenu(&[]MenuItem{
				MenuItem{Label: "Add user"},
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

// MenuItem menu item
type MenuItem struct {
	Label           string
	IsShow          func(*MenuItem) bool
	GetSubMenu      func(*MenuItem) *[]MenuItem
	SelectedFunc    func(*MenuItem) error
	backOptionLabel string
}

// ShowMenu show menu
func ShowMenu(label string, menu *[]MenuItem, backOptionLabel string) {
	for {
		menus := make([]string, 0)
		logger.Logger.Debugf("Show menu %+v", *menu)
		for _, menuItem := range *menu {
			if menuItem.IsShow == nil || menuItem.IsShow(&menuItem) {
				menus = append(menus, menuItem.Label)
			}
		}

		menus = append(menus, backOptionLabel)
		backIndex := len(menus) - 1

		menuPui := promptui.Select{
			Label: label,
			Items: menus,
		}

		index, subMenuLabel, err := menuPui.Run()
		logger.Logger.Debugf("Selected index: %d subMenuLabel: %+v err: %s", index, subMenuLabel, err)
		if err != nil {
			fmt.Printf("Select menu error %s\n", err)
			return
		}

		if index == backIndex {
			return
		}

		selected := (*menu)[index]

		if selected.GetSubMenu != nil {
			getSubMenu := selected.GetSubMenu
			subMenu := getSubMenu(&selected)
			back := "back"
			if selected.backOptionLabel != "" {
				back = selected.backOptionLabel
			}
			ShowMenu(subMenuLabel, subMenu, back)
		}

		if selected.SelectedFunc != nil {
			selectedFunc := selected.SelectedFunc
			err := selectedFunc(&selected)
			if err != nil {
				fmt.Printf("Gortal got an error %s\n", err)
			}
		}
	}
}

// ShowMainMenu show main menu
func ShowMainMenu() {
	ShowMenu("Please select the function you need", MainMenu, "Quit")
}
