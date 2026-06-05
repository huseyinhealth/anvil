package cmd

import (
	"anvil/internal"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func buildClasspath(verInfo *internal.VersionInfo, fabricProfile *internal.FabricProfile, instanceDir string) string {
    var paths []string

    for _, lib := range verInfo.Libraries {
        paths = append(paths, filepath.Join(instanceDir, "libraries", lib.Downloads.Artifact.Path))
    }

    for _, lib := range fabricProfile.Libraries {
        paths = append(paths, filepath.Join(instanceDir, "libraries", mavenPath(lib.Name)))
    }

    // minecraft jar en sona
    paths = append(paths, filepath.Join(instanceDir, "versions", verInfo.ID, verInfo.ID+".jar"))

    return strings.Join(paths, ":")
}

func parseArguments(args []json.RawMessage) []string {
    var result []string

    for _, arg := range args {
        // önce string olarak dene
        var s string
        if json.Unmarshal(arg, &s) == nil {
            result = append(result, s)
            continue
        }

        // string değilse koşullu argüman
        var conditional internal.ConditionalArg
        if json.Unmarshal(arg, &conditional) != nil {
            continue
        }

        // rule kontrolü
        allowed := false
        for _, rule := range conditional.Rules {
			if rule.Action == "allow" {
				if rule.Features != nil {
					allowed = false
					break
				}
				if rule.OS == nil || rule.OS.Name == mojangOS() {
					allowed = true
				}
			}
		}

        if !allowed {
            continue
        }

        // value string ya da []string olabilir
        var single string
        if json.Unmarshal(conditional.Value, &single) == nil {
            result = append(result, single)
            continue
        }

        var multiple []string
        if json.Unmarshal(conditional.Value, &multiple) == nil {
            result = append(result, multiple...)
        }
    }

    return result
}

func fillArgs(args []string, vars map[string]string) []string {
    result := make([]string, len(args))
    for i, arg := range args {
        for k, v := range vars {
            arg = strings.ReplaceAll(arg, "${"+k+"}", v)
        }
        result[i] = arg
    }
    return result
}

func Run(args... string) {
	buff, err := os.ReadFile(filepath.Join(internal.AnvilHome, ".selected"))

	if err != nil {
		fmt.Println("Use \"anvil select <instance name>\" first!")
		os.Exit(1)
	}

	instanceName := string(buff)
	
	
	var instance internal.Instance
	
	data, err := os.ReadFile(filepath.Join(internal.AnvilHome, "instances", instanceName, "anvil.json"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	json.Unmarshal(data, &instance)
	
	fmt.Printf("Starting instance %s \nVersion: %s\nFabric Version: %s\n",
		instance.Name, instance.Version, instance.FabricProfile.LoaderVersion);
	
	profile, e := GetProfile()

	if e > 0 {
		fmt.Println("Error: Couldn't start the game. Are you logged in?")
		os.Exit(1)
	}

	id, name, accessToken := profile.ID, profile.Name, profile.AccessToken
	version := instance.Version

	var versionJson internal.VersionInfo

	data, err = os.ReadFile(filepath.Join(internal.AnvilHome, "instances", instanceName, "versions", version, version+".json"))

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	json.Unmarshal(data, &versionJson)

	fMainClass := instance.FabricProfile.MainClass

	classPath := buildClasspath(&versionJson, &instance.FabricProfile, filepath.Join(internal.AnvilHome, "instances", instanceName))
	
	tmpDir, err := os.MkdirTemp("", "anvil-natives-*")
	/* TODO
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	for _, lib := range versionJson.Libraries {
		if lib.Downloads.Natives != nil {
			nativePath := filepath.Join(internal.AnvilHome, "instances", instanceName, "libraries", lib.Downloads.Natives.Path)
			extractTarGz(nativePath, tmpDir)
		}
	}
	*/

	
	argsGame := parseArguments(versionJson.Arguments.Game)
	argsJVM := parseArguments(versionJson.Arguments.JVM)

	vars := map[string]string{
		"auth_player_name":    name,
		"version_name":        version,
		"game_directory":      filepath.Join(internal.AnvilHome, "instances", instanceName),
		"assets_root":         filepath.Join(internal.AnvilHome, "instances", instanceName, "assets"),
		"assets_index_name":   versionJson.AssetIndex.ID,
		"auth_uuid":           id,
		"auth_access_token":   accessToken,
		"clientid":            internal.LauncherClientID,
		"auth_xuid":           "",
		"user_type":           "msa",
		"version_type":        versionJson.Type,
		"natives_directory":   tmpDir,
		"launcher_name":       "anvil",
		"launcher_version":    "1.0",
		"classpath":           classPath,
		"resolution_width":  "854",
		"resolution_height": "480",
		"quickPlayPath": "",
		"quickPlaySingleplayer": "",
		"quickPlayMultiplayer": "",
		"quickPlayRealms": "",
	}

	argsGame = fillArgs(argsGame, vars)
	argsJVM = fillArgs(argsJVM, vars)

	jreBase := filepath.Join(internal.JREPath, strconv.Itoa(versionJson.JavaVersion.MajorVersion))

	entries, err := os.ReadDir(jreBase)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	javaPath := filepath.Join(jreBase, entries[0].Name(), "bin", "java")

	cmdArgs := append(argsJVM, fMainClass)
	cmdArgs = append(cmdArgs, argsGame...)

	cmd := exec.Command(javaPath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = filepath.Join(internal.AnvilHome, "instances", instanceName)

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[ANVIL] bye bye")
}
