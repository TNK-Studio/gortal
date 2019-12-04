package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/TNK-Studio/gortal/src/config"
	"github.com/TNK-Studio/gortal/src/pui"
	"github.com/manifoldco/promptui"
)

var configPath = flag.String("c", fmt.Sprintf("%s%s", os.Getenv("HOME"), "/.gortal.yml"), "Config file")

func SubPromptSelect(prompt promptui.Select, itemsMap *map[string][]string) {
	for {
		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		if result == "quit" {
			break
		}

		if itemsMap != nil {
			subSelect := promptui.Select{
				Label: "Select Day",
				Items: (*itemsMap)[result],
			}
			SubPromptSelect(subSelect, nil)
		} else {
			break
		}
	}
}

func setupConfig() {
	fmt.Println("Config file not found. Setup config.", *config.ConfPath)
	_, _, err := pui.CreateUser(false, true)
	if err != nil {
		return
	}
	serverKey, _, err := pui.AddServer()
	if err != nil {
		return
	}
	_, _, err = pui.AddServerSSHUser(*serverKey)
	if err != nil {
		return
	}
	config.Conf.SaveTo(*config.ConfPath)
}

func main() {
	flag.Parse()
	if *configPath == "" {
		fmt.Println("Please specify a config file.")
		return
	}
	fmt.Println("Read config file", *configPath)
	config.ConfPath = configPath
	if !config.ConfigFileExisted(*config.ConfPath) {
		setupConfig()
	} else {
		config.Conf.ReadFrom(*config.ConfPath)
		server1 := (*config.Conf.Servers)["server1"]
		fmt.Println(server1)
	}

	// items1 := []string{"1", "2", "3"}
	// items1Map := &map[string][]string{
	// 	"1": []string{"1.1"},
	// 	"2": []string{"2.1"},
	// 	"3": []string{"3.1"},
	// }
	// prompt := promptui.Select{
	// 	Label: "Select Day",
	// 	Items: items1,
	// }

	// SubPromptSelect(prompt, items1Map)
}
