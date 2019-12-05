package pui

import (
	"fmt"

	"github.com/TNK-Studio/gortal/src/core/sshd"
	"github.com/TNK-Studio/gortal/src/utils/logger"
	"github.com/elfgzp/promptui"
	"github.com/gliderlabs/ssh"
)

// PUI pui
type PUI struct {
	sess *ssh.Session
}

// SetSession SetSession
func (ui *PUI) SetSession(s *ssh.Session) {
	ui.sess = s
}

// ShowMenu show menu
func (ui *PUI) ShowMenu(label string, menu *[]MenuItem, backOptionLabel string) {
	for {
		menuLabels := make([]string, 0)
		menuItems := make([]MenuItem, 0)

		if menu == nil {
			return
		}

		for index, menuItem := range *menu {
			if menuItem.IsShow == nil || menuItem.IsShow(index, &menuItem, ui.sess) {
				menuLabels = append(menuLabels, menuItem.Label)
				menuItems = append(menuItems, menuItem)
			}
		}
		logger.Logger.Debugf("Show menu %s", menuItems)

		menuLabels = append(menuLabels, backOptionLabel)
		backIndex := len(menuLabels) - 1
		menuPui := promptui.Select{
			Label:  label,
			Items:  menuLabels,
			Stdin:  *ui.sess,
			Stdout: *ui.sess,
		}

		index, subMenuLabel, err := menuPui.Run()

		logger.Logger.Debugf("Selected index: %d subMenuLabel: %+v err: %s", index, subMenuLabel, err)
		if err != nil {
			fmt.Printf("Select menu error %s\n", err)
			return
		}

		if index == backIndex {
			break
		}
		selected := &(menuItems[index])

		logger.Logger.Debugf("Selected: %+v", selected)

		if selected.GetSubMenu != nil {

			getSubMenu := selected.GetSubMenu
			subMenu := getSubMenu(index, selected, ui.sess)

			if subMenu != nil && len(*subMenu) > 0 {
				back := "back"
				if selected.backOptionLabel != "" {
					back = selected.backOptionLabel
				}

				if selected.SubMenuTitle != "" {
					subMenuLabel = selected.SubMenuTitle
				}
				ui.ShowMenu(subMenuLabel, subMenu, back)
			}
		}

		if selected.SelectedFunc != nil {
			selectedFunc := selected.SelectedFunc
			logger.Logger.Debugf("Run selectFunc %+v", selectedFunc)
			err := selectedFunc(index, selected, ui.sess)
			if err != nil {
				logger.Logger.Errorf("Run selected func err: %s", err)
				sshd.ErrorInfo(err, ui.sess)
			}
		}
	}
}

// ShowMainMenu show main menu
func (ui *PUI) ShowMainMenu() {
	ui.ShowMenu("Please select the function you need", MainMenu, "Quit")
}
