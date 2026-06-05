package internal

import "encoding/json"

type FabricVersion struct {
	Version string `json:"version"`
	Stable bool `json:"stable"`
}

type FabricLoaderVersion struct {
    Loader struct {
        Version string `json:"version"`
    } `json:"loader"`
}

type FabricProfile struct {
    MainClass string          `json:"mainClass"`
    Libraries []FabricLibrary `json:"libraries"`
	LoaderVersion string
}

type FabricLibrary struct {
    Name string `json:"name"`
    URL  string `json:"url"`
}

type VersionManifest struct {
    Versions []VersionEntry    `json:"versions"`
}

type VersionEntry struct {
    ID          string `json:"id"`
    Type        string `json:"type"`
    URL         string `json:"url"`
}

type VersionInfo struct {
    MainClass            string          `json:"mainClass"`
    Assets               string          `json:"assets"`
    AssetIndex           AssetIndex      `json:"assetIndex"`
    Downloads            ClientDownloads `json:"downloads"`
    Libraries            []Library       `json:"libraries"`
    JavaVersion          JavaVersion     `json:"javaVersion"`
    Arguments            Arguments       `json:"arguments"`
    Logging              Logging         `json:"logging"`
    ReleaseTime          string          `json:"releaseTime"`
    Time                 string          `json:"time"`
    Type                 string          `json:"type"`
    MinimumLauncherVersion int           `json:"minimumLauncherVersion"`
    ComplianceLevel      int             `json:"complianceLevel"`
    ID                   string          `json:"id"`
}

type AssetIndex struct {
    ID         string `json:"id"`
    URL        string `json:"url"`
    Sha1       string `json:"sha1"`
    Size       int    `json:"size"`
    TotalSize  int    `json:"totalSize"`
	Objects map[string]Asset `json:"objects"`
}

type ClientDownloads struct {
    Client Download `json:"client"`
    Server Download `json:"server"`
}

type Download struct {
    URL  string `json:"url"`
    Sha1 string `json:"sha1"`
    Size int    `json:"size"`
}

type Library struct {
    Name      string          `json:"name"`
    Downloads LibraryDownload `json:"downloads"`
    Rules     []Rule          `json:"rules,omitempty"`
}

type LibraryDownload struct {
    Artifact Artifact `json:"artifact"`
}

type Artifact struct {
    Path string `json:"path"`
    URL  string `json:"url"`
    Sha1 string `json:"sha1"`
    Size int    `json:"size"`
}

type Rule struct {
    Action string   `json:"action"`
    OS     *OS      `json:"os,omitempty"`
}

type OS struct {
    Name string `json:"name"`
}

type JavaVersion struct {
    Component    string `json:"component"`
    MajorVersion int    `json:"majorVersion"`
}

type Arguments struct {
    Game []json.RawMessage `json:"game"`
    JVM  []json.RawMessage `json:"jvm"`
}

type ConditionalArg struct {
    Rules []ArgRule `json:"rules"`
    Value json.RawMessage `json:"value"`
}

type ArgRule struct {
    Action   string              `json:"action"`
    OS       *OS                 `json:"os,omitempty"`
    Features map[string]bool     `json:"features,omitempty"`
}

type Logging struct {
    Client LoggingClient `json:"client"`
}

type LoggingClient struct {
    Argument string      `json:"argument"`
    File     LoggingFile `json:"file"`
    Type     string      `json:"type"`
}

type LoggingFile struct {
    ID   string `json:"id"`
    URL  string `json:"url"`
    Sha1 string `json:"sha1"`
    Size int    `json:"size"`
}

type Asset struct {
    Hash string `json:"hash"`
    Size int    `json:"size"`
}

type DeviceCodeResponse struct {
    DeviceCode      string `json:"device_code"`
    UserCode        string `json:"user_code"`
    VerificationURI string `json:"verification_uri"`
    ExpiresIn       int    `json:"expires_in"`
    Interval        int    `json:"interval"`
}

type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
    Error        string `json:"error"`
}

type XBLResponse struct {
    Token         string `json:"Token"`
    DisplayClaims struct {
        Xui []struct {
            Uhs string `json:"uhs"`
        } `json:"xui"`
    } `json:"DisplayClaims"`
}

type XSTSResponse struct {
    Token         string `json:"Token"`
    DisplayClaims struct {
        Xui []struct {
            Uhs string `json:"uhs"`
        } `json:"xui"`
    } `json:"DisplayClaims"`
}

type MinecraftTokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}

type MinecraftProfile struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    AccessToken string `json:"accessToken"`
}

type ModrinthProject struct {
    ID   string `json:"id"`
    Slug string `json:"slug"`
    Title string `json:"title"`
    Description string `json:"description"`
    ClientSide string `json:"client_side"`
    ServerSide string `json:"server_side"`
}

type ModrinthSearch struct {
    Hits []ModrinthProject `json:"hits"`
}

type ModrinthVersion struct {
    ID           string                `json:"id"`
    VersionNumber string               `json:"version_number"`
    Files        []ModrinthFile        `json:"files"`
    Dependencies []ModrinthDependency  `json:"dependencies"`
}

type ModrinthFile struct {
    URL      string `json:"url"`
    Filename string `json:"filename"`
    Primary  bool   `json:"primary"`
    Size int64 `json:"size"`
}

type ModrinthDependency struct {
    ProjectID   string `json:"project_id"`
    VersionID   string `json:"version_id"`
    DependencyType string `json:"dependency_type"`
}
