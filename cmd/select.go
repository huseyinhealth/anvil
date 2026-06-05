package cmd

import (
	"anvil/internal"
	"fmt"
	"os"
	"path/filepath"
)

func Select(args ...string) {
    name := args[0]
    instancePath := filepath.Join(internal.InstancesPath, name)

    if _, err := os.Stat(instancePath); os.IsNotExist(err) {
        fmt.Printf("Error: Instance '%s' does not exist.\n", name)
        os.Exit(1)
    }

    selectedPath := filepath.Join(internal.AnvilHome, ".selected")
    err := os.WriteFile(selectedPath, []byte(name), 0644)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Selected instance: %s\n", name)
}
