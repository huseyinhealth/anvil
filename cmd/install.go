package cmd

import (
	"anvil/internal"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func getProjectBySlug(slug string) (internal.ModrinthProject, error) {
    var project internal.ModrinthProject
    resp, err := rClient.R().
        SetResult(&project).
        Get("https://api.modrinth.com/v2/project/" + slug)
    if err != nil || resp.IsError() {
        return project, fmt.Errorf("project not found: %s", slug)
    }
    return project, nil
}

func getVersions(projectID, gameVersion string) ([]internal.ModrinthVersion, error) {
    var versions []internal.ModrinthVersion
    resp, err := rClient.R().
        SetResult(&versions).
        SetQueryParams(map[string]string{
            "game_versions": `["` + gameVersion + `"]`,
            "loaders":       `["fabric"]`,
        }).
        Get("https://api.modrinth.com/v2/project/" + projectID + "/version")
    if err != nil || resp.IsError() {
        return nil, fmt.Errorf("versions not found for: %s", projectID)
    }
    return versions, nil
}

type ModInstallPlan struct {
    Mod        internal.Mod
    FileURL    string
    Filename   string
    Size       int64
    RequiredBy string
	AlreadyInstalled bool
}

func resolveMod(slug, gameVersion string, installedMods []internal.Mod) ([]ModInstallPlan, error) {
    project, err := getProjectBySlug(slug)
    if err != nil {
        return nil, err
    }

    for i, installed := range installedMods {
        if installed.ProjectID == project.ID {
            return []ModInstallPlan{{
				Mod:        installedMods[i],
				AlreadyInstalled: true,
			}}, nil
        }
    }

    versions, err := getVersions(project.ID, gameVersion)
    if err != nil || len(versions) == 0 {
        return nil, fmt.Errorf("no compatible version found for: %s", slug)
    }

    version := versions[0]

    incompatible kontrolü
    for _, dep := range version.Dependencies {
        if dep.DependencyType == "incompatible" {
            for _, installed := range installedMods {
                if installed.ProjectID == dep.ProjectID {
                    return nil, fmt.Errorf("incompatible mod already installed: %s", installed.Name)
                }
            }
        }
    }

    var fileURL, filename string
    var size int64
    for _, f := range version.Files {
        if f.Primary {
            fileURL = f.URL
            filename = f.Filename
            size = f.Size
            break
        }
    }

    mod := internal.Mod{
        Slug:      slug,
        Name:      project.Title,
        Version:   version.VersionNumber,
        ProjectID: project.ID,
        VersionID: version.ID,
    }

    plans := []ModInstallPlan{{Mod: mod, FileURL: fileURL, Filename: filename, Size: size}}

    for _, dep := range version.Dependencies {
        if dep.DependencyType != "required" {
            continue
        }
        depProject, err := getProjectBySlug(dep.ProjectID)
        if err != nil {
            continue
        }
        depPlans, err := resolveMod(depProject.Slug, gameVersion, installedMods)
        if err != nil {
            return nil, err
        }
        for i := range depPlans {
            if depPlans[i].RequiredBy == "" {
                depPlans[i].RequiredBy = slug
            }
        }
        plans = append(plans, depPlans...)
    }

    return plans, nil
}

func executePlans(plans []ModInstallPlan, modsDir string, instance *internal.Instance) ([]internal.Mod, error) {
    var newMods []internal.Mod
    var downloadedFiles []string

    for _, plan := range plans {
        if plan.AlreadyInstalled {
            if plan.RequiredBy != "" {
                for i := range instance.Mods {
                    if instance.Mods[i].ProjectID == plan.Mod.ProjectID {
                        instance.Mods[i].RequiredBy = append(instance.Mods[i].RequiredBy, plan.RequiredBy)
                        break
                    }
                }
                for i := range newMods {
                    if newMods[i].ProjectID == plan.Mod.ProjectID {
                        newMods[i].RequiredBy = append(newMods[i].RequiredBy, plan.RequiredBy)
                        break
                    }
                }
            }
            continue
        }

        destPath := filepath.Join(modsDir, plan.Filename)
        err := downloadFile(plan.FileURL, destPath, "")
        if err != nil {
            // rollback
            fmt.Printf("\nError installing %s, rolling back...\n", plan.Mod.Name)
            for _, f := range downloadedFiles {
                os.Remove(f)
            }
            return nil, fmt.Errorf("installation failed: %v", err)
        }

        downloadedFiles = append(downloadedFiles, destPath)

        if plan.RequiredBy != "" {
            plan.Mod.RequiredBy = append(plan.Mod.RequiredBy, plan.RequiredBy)
        }
        newMods = append(newMods, plan.Mod)
    }

    return newMods, nil
}

func plannedMods(plans []ModInstallPlan) []internal.Mod {
    var mods []internal.Mod
    for _, p := range plans {
        mods = append(mods, p.Mod)
    }
    return mods
}

func Install(args ...string) {
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

    rClient = internal.NewClient()
	
    modsDir := filepath.Join(instanceDir, "mods")

    var allPlans []ModInstallPlan

	fabricAPIInstalled := false
	for _, mod := range instance.Mods {
		if mod.Slug == "fabric-api" {
			fabricAPIInstalled = true
			break
		}
	}
	if !fabricAPIInstalled {
		fabricPlans, err := resolveMod("fabric-api", instance.Version, instance.Mods)
		if err == nil {
			allPlans = append(allPlans, fabricPlans...)
		}
	}

    for _, slug := range args {
        alreadyInstalled := false
        for _, installed := range instance.Mods {
            if installed.Slug == slug {
                fmt.Printf("%s is already installed.\n", slug)
                alreadyInstalled = true
                break
            }
        }
        if alreadyInstalled {
            continue
        }
			
        tempMods := append(instance.Mods, plannedMods(allPlans)...)
        plans, err := resolveMod(slug, instance.Version, tempMods)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            os.Exit(1)
        }
        allPlans = append(allPlans, plans...)
    }

    if len(allPlans) == 0 {
        fmt.Println("Nothing to install.")
        return
    }

    var totalSize int64
	var installCount int
	fmt.Println("The following mods will be installed:")
	for _, p := range allPlans {
		if p.AlreadyInstalled {
			continue
		}
		fmt.Printf("  %-30s %s (%.1f MB)\n", p.Mod.Name, p.Mod.Version, float64(p.Size)/1024/1024)
		totalSize += p.Size
		installCount++
	}
    fmt.Printf("\nTotal: %d mods, %.1f MB\n", installCount, float64(totalSize)/1024/1024)
    fmt.Print("Proceed? [Y/n] ")

    var input string
    fmt.Scanln(&input)
    if input == "n" || input == "N" {
        fmt.Println("Aborted.")
        return
    }

    mods, err := executePlans(allPlans, modsDir, &instance)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

    for _, mod := range mods {
        instance.Mods = append(instance.Mods, mod)
    }

	for i := range allPlans {
		for j := range allPlans {
			if i == j {
				continue
			}
			if allPlans[i].RequiredBy == allPlans[j].Mod.Slug {
				allPlans[j].Mod.RequiredBy = append(allPlans[j].Mod.RequiredBy, allPlans[i].RequiredBy)
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

    fmt.Printf("\nInstalled %d mod(s).\n", len(mods))
}
