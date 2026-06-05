package cmd

import (
	"anvil/internal"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Remove(args ...string) {
    buff, err := os.ReadFile(filepath.Join(internal.AnvilHome, ".selected"))
    if err != nil {
        fmt.Println("use \"anvil select <instance name>\" first!")
        os.Exit(1)
    }
    instanceName := string(buff)
    instanceDir := filepath.Join(internal.InstancesPath, instanceName)

    var instance internal.Instance
    data, err := os.ReadFile(filepath.Join(instanceDir, "anvil.json"))
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    json.Unmarshal(data, &instance)

    // kaldırılacak modları bul ve kontrol et
    var toRemove []internal.Mod
    for _, slug := range args {
        found := false
        for _, mod := range instance.Mods {
            if mod.Slug == slug {
                // başka mod bu modu gerektiriyor mu
                if len(mod.RequiredBy) > 0 {
                    fmt.Printf("Error: %s is required by %s\n", slug, strings.Join(mod.RequiredBy, ", "))
                    os.Exit(1)
                }
				alreadyInList := false
				for _, r := range toRemove {
					if r.Slug == slug {
						alreadyInList = true
						break
					}
				}
				if !alreadyInList {
					toRemove = append(toRemove, mod)
				}
                found = true
                break
            }
        }
        if !found {
            fmt.Printf("Error: %s is not installed.\n", slug)
            os.Exit(1)
        }
    }

    // onay ekranı
    fmt.Println("The following mods will be removed:")
    for _, mod := range toRemove {
        fmt.Printf("  %s %s\n", mod.Name, mod.Version)
    }
    fmt.Printf("\nTotal: %d mod(s)\n", len(toRemove))
    fmt.Print("Proceed? [Y/n] ")

    var input string
    fmt.Scanln(&input)
    if input == "n" || input == "N" {
        fmt.Println("Aborted.")
        return
    }

    // jar'ları sil ve instance.Mods'dan çıkar
    for _, mod := range toRemove {
        // mods klasöründeki jar'ı bul ve sil
        modsDir := filepath.Join(instanceDir, "mods")
        entries, _ := os.ReadDir(modsDir)
        for _, entry := range entries {
            if strings.Contains(entry.Name(), mod.Slug) {
                os.Remove(filepath.Join(modsDir, entry.Name()))
            }
        }

        // instance.Mods'dan çıkar
        for i, m := range instance.Mods {
            if m.Slug == mod.Slug {
                instance.Mods = append(instance.Mods[:i], instance.Mods[i+1:]...)
                break
            }
        }

        // requiredBy güncellemesi
        for i := range instance.Mods {
            for j, rb := range instance.Mods[i].RequiredBy {
                if rb == mod.Slug {
                    instance.Mods[i].RequiredBy = append(instance.Mods[i].RequiredBy[:j], instance.Mods[i].RequiredBy[j+1:]...)
                    break
                }
            }
        }
    }

    data, err = json.MarshalIndent(instance, "", "  ")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    err = os.WriteFile(filepath.Join(instanceDir, "anvil.json"), data, 0644)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("\nRemoved %d mod(s).\n", len(toRemove))
}
