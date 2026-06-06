package internal

type Command struct {
	Name string
	Description string
	Alias []string
	MinArgs int
	F func(...string)
}

type Mod struct {
    Slug       string   `json:"slug"`
    Name       string   `json:"name"`
    Version    string   `json:"version"`
    ProjectID  string   `json:"projectId"`
    VersionID  string   `json:"versionId"`
    RequiredBy []string `json:"requiredBy"`
}

type Instance struct {
    Name    string `json:"name"`
    Version string `json:"version"`
	LoaderVersion string `json:"loaderVersion"`
	MainClass string `json:"mainClass"`
	FabricProfile FabricProfile `json:"fabricProfile"`
    Mods    []Mod  `json:"mods"`
}

type HashedFile struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
	URL string `json:"url"`
}
