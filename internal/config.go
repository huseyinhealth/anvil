package internal

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

var AnvilHome string
var InstancesPath string
var JREPath string

var FileCache string

var LauncherClientID string

func Init() {
	home, _ := os.UserHomeDir()
	AnvilHome = filepath.Join(home, ".anvil")
	InstancesPath = filepath.Join(AnvilHome, "instances")
	JREPath = filepath.Join(AnvilHome, "jre")

	FileCache = filepath.Join(AnvilHome, "filecache.json")

	godotenv.Load()
    if id := os.Getenv("LAUNCHER_CLIENT_ID"); id != "" {
        LauncherClientID = id
    }
}
