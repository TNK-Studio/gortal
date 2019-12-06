package pui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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
