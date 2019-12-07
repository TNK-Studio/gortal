package pui

import (
	"fmt"
	"io"
	"strings"

	"github.com/TNK-Studio/gortal/config"
	"github.com/elfgzp/promptui"
)

var serverNamePrompt = func(defaultShow string, stdio io.ReadWriteCloser) promptui.Prompt {
	return promptui.Prompt{
		Label:    "Server name",
		Validate: Required("server name"),
		Default:  defaultShow,
		Stdin:    stdio,
		Stdout:   stdio,
	}
}

var serverHostPrompt = func(defaultShow string, stdio io.ReadWriteCloser) promptui.Prompt {
	return promptui.Prompt{
		Label:    "Server host",
		Validate: Required("server host"),
		Default:  defaultShow,
		Stdin:    stdio,
		Stdout:   stdio,
	}
}

var serverPortPrompt = func(defaultShow string, stdio io.ReadWriteCloser) promptui.Prompt {
	return promptui.Prompt{
		Label: "Server port",
		Validate: MultiValidate(
			[](func(string) error){
				Required("server port"),
				IsInt(),
			},
		),
		Default: defaultShow,
		Stdin:   stdio,
		Stdout:  stdio,
	}
}

var sshUserNamePrompt = func(defaultShow string, stdio io.ReadWriteCloser) promptui.Prompt {
	return promptui.Prompt{
		Label:    "SSH username",
		Validate: Required("SSH username"),
		Default:  defaultShow,
		Stdin:    stdio,
		Stdout:   stdio,
	}
}

var identityFilePrompt = func(label, defaultShow string, stdio io.ReadWriteCloser, validates ...func(string) error) promptui.Prompt {
	return promptui.Prompt{
		Label: label,
		Validate: MultiValidate(
			append(
				[](func(string) error){
					Required("identity file path"),
				},
				validates...,
			),
		),
		Default: defaultShow,
		Stdin:   stdio,
		Stdout:  stdio,
	}
}

var allowAllUserPrompt = func(defaultShow string, stdio io.ReadWriteCloser) promptui.Prompt {
	return promptui.Prompt{
		Label:    "Allow all user access ? yes/no",
		Validate: YesOrNo(),
		Default:  defaultShow,
		Stdin:    stdio,
		Stdout:   stdio,
	}
}

var allowUsersPrompt = func(defaultShow string, stdio io.ReadWriteCloser) promptui.Prompt {
	return promptui.Prompt{
		Label:   "Please enter all usernames separated by ','",
		Default: defaultShow,
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
							return fmt.Errorf(
								"Username '%s' of user not existed. Please choose from %v. ",
								username,
								userList,
							)
						}
					}
					return nil
				},
			},
		),
		Stdin:  stdio,
		Stdout: stdio,
	}
}
