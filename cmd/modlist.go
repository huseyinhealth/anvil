package cmd

import (
	"anvil/internal"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

func ModList(args... string) {

	var instance internal.Instance

	instanceName, er := internal.GetSelectedInstance()

	if er != 0 {
		fmt.Println("Error: Use \"anvil select <insanceName>\" first!")
		os.Exit(1)
	}

	data, err := os.ReadFile(filepath.Join(internal.AnvilHome, "instances", instanceName, "anvil.json"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	json.Unmarshal(data, &instance)

	if len(instance.Mods) == 0 {
		fmt.Println("You didn't install any mods!")
		return
	}

	// FgGreen yerine FgHiGreen (Bright Green) kullanıyoruz
    brightGreen := color.New(color.FgHiGreen).SprintFunc()

    fmt.Printf("Installed mods on instance \"%s\":\n\n", instanceName)

    for _, i := range instance.Mods {
        // Artık tam APT tarzı açık yeşil basacak
        fmt.Printf("%s - %s\n", brightGreen(i.Slug), i.Name)
    }
}