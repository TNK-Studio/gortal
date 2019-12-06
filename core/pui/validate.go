package pui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/TNK-Studio/gortal/utils"
)

// MultiValidate MultiValidate
func MultiValidate(validates [](func(string) error)) func(string) error {
	return func(input string) error {
		for _, validateFunc := range validates {
			err := validateFunc(input)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// Required required validate
func Required(field string) func(string) error {
	return func(input string) error {
		if strings.ReplaceAll(input, " ", "") == "" {
			return fmt.Errorf("Please input %s", field)
		}
		return nil
	}
}

// IsInt check input
func IsInt() func(string) error {
	return func(input string) error {
		if _, err := strconv.Atoi(input); err != nil {
			return errors.New("Please check if your input is an integer")
		}
		return nil
	}
}

// YesOrNo check yes/no
func YesOrNo() func(string) error {
	return func(input string) error {
		if input == "yes" || input == "no" {
			return nil
		}
		return errors.New("Please input yes / no")
	}
}

// FileExited FileExited
func FileExited(filename string) func(string) error {
	return func(input string) error {
		if !utils.FileExited(utils.FilePath(input)) {
			return fmt.Errorf("%s '%s' not existed", filename, input)
		}
		return nil
	}
}

// FileNotExited FileNotExited
func FileNotExited(filename string) func(string) error {
	return func(input string) error {
		if utils.FileExited(utils.FilePath(input)) {
			return fmt.Errorf("%s '%s' existed", filename, input)
		}
		return nil
	}
}

// IsDir IsDir
func IsDir() func(string) error {
	return func(input string) error {
		if !utils.IsDirector(utils.FilePath(input)) {
			return fmt.Errorf("Path '%s' is not a director", input)
		}
		return nil
	}
}

// IsNotDir IsNotDir
func IsNotDir() func(string) error {
	return func(input string) error {
		if utils.IsDirector(utils.FilePath(input)) {
			return fmt.Errorf("Path '%s' is a director", input)
		}
		return nil
	}
}
