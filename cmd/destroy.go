package cmd

import (
	"anvil/internal"
	"fmt"
	"os"
	"path/filepath"
)

func Destroy(args... string) {

	instanceName := args[0]

	_, err := os.ReadDir(filepath.Join(internal.InstancesPath, instanceName))

	if err != nil {
		fmt.Printf("Error: No instance named \"%s\" found.\n", instanceName)
		os.Exit(1)
	}

	fmt.Printf("This will permanently delete instance \"%s\". Proceed? [y/N] ", instanceName)
    var input string
    fmt.Scanln(&input)
    if input != "y" && input != "Y" {
        fmt.Println("Aborted.")
        return
    }

	if n, e := internal.GetSelectedInstance(); n == instanceName && e == 0 {
		err := os.Remove(filepath.Join(internal.AnvilHome, ".selected"))

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	err = os.RemoveAll(filepath.Join(internal.InstancesPath, instanceName))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}