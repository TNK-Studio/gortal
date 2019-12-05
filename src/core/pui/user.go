package pui

import (
	"errors"
	"fmt"

	"github.com/TNK-Studio/gortal/src/config"
	"github.com/elfgzp/promptui"
)

// CreateUser new user
func CreateUser(showAdminSelect bool, isAdmin bool) (*string, *config.User, error) {
	fmt.Println("Create a user.")
	usernamePui := promptui.Prompt{
		Label: "Username",
		Validate: func(input string) error {
			if len(input) < 3 {
				return errors.New("Username must have more than 3 characters")
			}
			return nil
		},
		Stdin:  Sess,
		Stdout: Sess,
	}

	username, err := usernamePui.Run()
	if err != nil {
		return nil, nil, err
	}

	passwdPui := promptui.Prompt{
		Label: "Password",
		Validate: func(input string) error {
			if len(input) < 6 {
				return errors.New("Password must have more than 6 characters")
			}
			return nil
		},
		Mask:   '*',
		Stdin:  Sess,
		Stdout: Sess,
	}

	passwd, err := passwdPui.Run()
	if err != nil {
		return nil, nil, err
	}

	confirmPasswdPui := promptui.Prompt{
		Label: "Confirm your password",
		Validate: func(input string) error {
			if input != passwd {
				return errors.New("Password not match")
			}
			return nil
		},
		Mask:   '*',
		Stdin:  Sess,
		Stdout: Sess,
	}

	_, err = confirmPasswdPui.Run()
	if err != nil {
		return nil, nil, err
	}

	IsAdminString := ""
	if showAdminSelect && !isAdmin {
		adminPui := promptui.Prompt{
			Label:    "Is admin ? yes/no",
			Validate: YesOrNo(),
			Stdin:    Sess,
			Stdout:   Sess,
		}

		IsAdminString, err = adminPui.Run()
		if err != nil {
			return nil, nil, err
		}
	}

	isAdmin = IsAdminString == "yes" || isAdmin
	if isAdmin {
		fmt.Println("Create a admin user")
	}
	key, user := config.Conf.AddUser(username, passwd, isAdmin)
	return &key, user, nil
}
