package pui

import (
	"fmt"

	"github.com/TNK-Studio/gortal/src/core/state"
	"github.com/TNK-Studio/gortal/src/utils/logger"
	"github.com/elfgzp/promptui"
)

// MainMenu main menu
var MainMenu *[]MenuItem

func defaultShow(int, *MenuItem) bool { return true }

func isAdmin(int, *MenuItem) bool { return state.CurrentUser.Admin }

func staticSubMenu(subMenu *[]MenuItem) func(int, *MenuItem) *[]MenuItem {
	return func(int, *MenuItem) *[]MenuItem {
		return subMenu
	}
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
	IsShow          func(int, *MenuItem) bool
	SubMenuTitle    string
	GetSubMenu      func(int, *MenuItem) *[]MenuItem
	SelectedFunc    func(int, *MenuItem) error
	backOptionLabel string
}

// ShowMenu show menu
func ShowMenu(label string, menu *[]MenuItem, backOptionLabel string) {
	for {
		menus := make([]string, 0)
		logger.Logger.Debugf("Show menu %+v", *menu)
		if menu == nil {
			return
		}

		for index, menuItem := range *menu {
			if menuItem.IsShow == nil || menuItem.IsShow(index, &menuItem) {
				menus = append(menus, menuItem.Label)
			}
		}

		menus = append(menus, backOptionLabel)
		backIndex := len(menus) - 1

		menuPui := promptui.Select{
			Label:  label,
			Items:  menus,
			Stdin:  Sess,
			Stdout: Sess,
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
		logger.Logger.Debugf("Selected: %+v", selected)

		if selected.GetSubMenu != nil {

			getSubMenu := selected.GetSubMenu
			subMenu := getSubMenu(index, &selected)

			if subMenu != nil && len(*subMenu) > 0 {
				back := "back"
				if selected.backOptionLabel != "" {
					back = selected.backOptionLabel
				}

				if selected.SubMenuTitle != "" {
					subMenuLabel = selected.SubMenuTitle
				}
				ShowMenu(subMenuLabel, subMenu, back)
			}

		}

		if selected.SelectedFunc != nil {
			selectedFunc := selected.SelectedFunc
			logger.Logger.Debugf("Run selectFunc %+v", selectedFunc)
			err := selectedFunc(index, &selected)
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
