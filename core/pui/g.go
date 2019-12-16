package pui

import (
	"errors"

	"github.com/TNK-Studio/gortal/core/sshd"
	"github.com/TNK-Studio/gortal/utils/logger"
	"github.com/elfgzp/promptui"
	"github.com/elfgzp/ssh"
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
func (ui *PUI) ShowMenu(label string, menu *[]*MenuItem, BackOptionLabel string, selectedChain []*MenuItem) {
	for {
		menuLabels := make([]string, 0)
		menuItems := make([]*MenuItem, 0)

		if menu == nil {
			break
		}

		for index, menuItem := range *menu {
			if menuItem.IsShow == nil || menuItem.IsShow(index, menuItem, ui.sess, selectedChain) {
				menuLabels = append(menuLabels, menuItem.Label)
				menuItems = append(menuItems, menuItem)
			}
		}
		logger.Logger.Debugf("Show menu %s", menuItems)

		menuLabels = append(menuLabels, BackOptionLabel)
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
			logger.Logger.Infof("Select menu error %s\n", err)
			break
		}

		if index == backIndex {
			break
		}

		selected := menuItems[index]

		logger.Logger.Debugf("Selected: %+v", selected)

		if selected.GetSubMenu != nil {

			getSubMenu := selected.GetSubMenu
			subMenu := getSubMenu(index, selected, ui.sess, selectedChain)

			if subMenu != nil && len(*subMenu) > 0 {
				back := "back"
				if selected.BackOptionLabel != "" {
					back = selected.BackOptionLabel
				}

				if selected.SubMenuTitle != "" {
					subMenuLabel = selected.SubMenuTitle
				}
				ui.ShowMenu(subMenuLabel, subMenu, back, append(selectedChain, selected))
			} else {
				noSubMenuInfo := "No options under this menu ... "
				if selected.NoSubMenuInfo != "" {
					noSubMenuInfo = selected.NoSubMenuInfo
				}
				sshd.ErrorInfo(errors.New(noSubMenuInfo), ui.sess)
			}
		}

		if selected.SelectedFunc != nil {
			selectedFunc := selected.SelectedFunc
			logger.Logger.Debugf("Run selectFunc %+v", selectedFunc)
			err := selectedFunc(index, selected, ui.sess, selectedChain)
			if err != nil {
				logger.Logger.Errorf("Run selected func err: %s", err)
				sshd.ErrorInfo(err, ui.sess)
			}
			if selected.BackAfterSelected == true {
				break
			}
		}
	}
}

// ShowMainMenu show main menu
func (ui *PUI) ShowMainMenu() {
	selectedChain := make([]*MenuItem, 0)
	ui.ShowMenu("Please select the function you need", MainMenu, "Quit", selectedChain)
}
