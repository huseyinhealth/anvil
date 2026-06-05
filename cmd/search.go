package cmd

import (
	"anvil/internal"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

func Search(args... string) {
	query := args[0]

	instanceFound := false

	instanceName, er := internal.GetSelectedInstance()

	if er == 0 {
		instanceFound = true
	}

	if !instanceFound {
		fmt.Println("Use \"anvil select <instanceName>\" first!")
		os.Exit(1)
	}

	data, err := os.ReadFile(filepath.Join(internal.AnvilHome, "instances", instanceName, "anvil.json"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

    var	instance internal.Instance
	json.Unmarshal(data, &instance)

	version := instance.Version
	loader := "fabric"
	facets := fmt.Sprintf("[[\"categories:%s\"],[\"versions:%s\"],[\"project_type:mod\"]]", loader, version)
	limit := 100
	offset := 0

	endpoint := "https://api.modrinth.com/v2/search"

    rClient := internal.NewClient()

	var result internal.ModrinthSearch

    // 2. Parametreleri SetQueryParams ile kütüphaneye devrediyoruz (Otomatik URL Encode yapar)
    _, err = rClient.R().
        SetQueryParams(map[string]string{
            "query":  query,
            "facets": facets,
            "limit":  fmt.Sprintf("%d", limit),
            "offset": fmt.Sprintf("%d", offset),
        }).
		SetResult(&result).
        Get(endpoint)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	green := color.New(color.FgGreen).SprintFunc()
    faint := color.New(color.Faint).SprintFunc()

    for _, i := range result.Hits {
        fmt.Printf("%s - %s\n", green(i.Slug), i.Title)
        fmt.Printf("    %s\n\n", faint(i.Description))
    }
}
