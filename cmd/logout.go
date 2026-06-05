package cmd

import (
	"anvil/internal"
	"fmt"
	"os"
	"path/filepath"
)

func Logout(...string) {
    profilePath := filepath.Join(internal.AnvilHome, "profile.json")
    
    if _, err := os.Stat(profilePath); os.IsNotExist(err) {
        fmt.Println("Not logged in.")
        os.Exit(1)
    }

    err := os.Remove(profilePath)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("Logged out.")
}
