package cmd

import (
	"anvil/internal"
	"fmt"
	"os"
)

func List(args... string) {
	fmt.Print("Instances list:\n")

	ins, err := os.ReadDir(internal.InstancesPath)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	for _, i := range ins {
		if i.IsDir() {
			fmt.Printf("%s\n", i.Name())
		}
	}
}