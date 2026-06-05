package cmd

import (
	"anvil/internal"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func GetProfile() (internal.MinecraftProfile, int) {
	_, err := os.Stat(filepath.Join(internal.AnvilHome, "profile.json"))

	if errors.Is(err, os.ErrNotExist) {
		return internal.MinecraftProfile{}, 1
	}

	var profile internal.MinecraftProfile

	data, err := os.ReadFile(filepath.Join(internal.AnvilHome, "profile.json"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return internal.MinecraftProfile{}, 2
	}
	json.Unmarshal(data, &profile)
	return profile, 0
}

func Status(args... string) {
	profile, error := GetProfile()

	var instance internal.Instance

	instanceFound := false

	instanceName, er := internal.GetSelectedInstance()

	if er == 0 {
		instanceFound = true
	}

	data, err := os.ReadFile(filepath.Join(internal.AnvilHome, "instances", instanceName, "anvil.json"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	json.Unmarshal(data, &instance)

	fmt.Println("- Anvil Status -")

	if error == 1 {
		fmt.Println("You are not logged-in.")
	} else {
		fmt.Printf("Logged in as: %s\n", profile.Name)
	}
	
	fmt.Printf("Selected instance: %s\n", instance.Name)

	if instanceFound {
		fmt.Println("- Instance Status -")
		fmt.Printf("Name: %v\n", instance.Name)
		fmt.Printf("Minecraft Version: %v\n", instance.Version)
		fmt.Printf("Forge Version: %v\n", instance.LoaderVersion)
		fmt.Printf("Installed Mods: %v\n", len(instance.Mods))
	}
}