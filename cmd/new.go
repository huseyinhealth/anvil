package cmd

import (
	"anvil/internal"
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mattn/go-runewidth"

	"github.com/go-resty/resty/v2"
	"golang.org/x/term"
)

const FM_VERSIONS_GAME = "https://meta.fabricmc.net/v1/versions/game/%s"
const FM_LOADER = "https://meta.fabricmc.net/v2/versions/loader/%s"
const FM_PROFILE = "https://meta.fabricmc.net/v2/versions/loader/%s/%s/profile/json"

const MM_V_MANIFEST = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
const A_JAVA = "https://api.adoptium.net/v3/binary/latest/%d/ga/linux/x64/jre/hotspot/normal/eclipse"

var rClient *resty.Client

var hashedFiles []internal.HashedFile
var cacheExists bool = true
var cachesMu sync.Mutex
var downloadMu sync.Map

type downloaded struct {
	id int
	finished bool
}

var downloadedFiles []*downloaded
var displayed int

type progressWriter struct {
	dest    string
	total   int64
	written int64
	file    *os.File
	last    time.Time
	dled *downloaded
	displaying int
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.file.Write(p)
	pw.written += int64(n)

	// 50ms'de bir güncelle
	if time.Since(pw.last) < 50*time.Millisecond && err == nil {
		return n, err
	}
	pw.last = time.Now()

	if pw.total > 0 {
		printProgress(filepath.Base(pw.dest), pw.written, pw.total, pw)
	}

	return n, err
}

func printProgress(name string, written, total int64, pw *progressWriter) {
	if pw.dled.id == 0 {
		return
	}

	displayed = pw.dled.id

	busy := false
	for _, i := range downloadedFiles {
		if !i.finished && i.id != pw.dled.id{
			busy = true
			

			break
			
		}
	}

	for _, i := range downloadedFiles {
		if i.id == displayed && pw.dled.id != i.id {
			
			busy = true
			break
		}
	}

	if busy {
		return
	}
	const barWidth = 30

	percent := float64(written) / float64(total)
	filled := int(percent * barWidth)

	bar := strings.Repeat("█", filled) +
		strings.Repeat("░", barWidth-filled)

	speed := fmt.Sprintf("%.1f MB",
		float64(written)/(1024*1024))

	msg := fmt.Sprintf(
		"%-20s [%s] %6.2f%% %s/%s",
		truncate(name, 20),
		bar,
		percent*100,
		speed,
		fmt.Sprintf("%.1f MB", float64(total)/(1024*1024)),
	)

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))

	if runewidth.StringWidth(msg) < width {
		msg += strings.Repeat(" ", width-runewidth.StringWidth(msg))
	}

	fmt.Printf("\r%s", msg)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func assetURL(hash string) string {
    return fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", hash[:2], hash)
}

func mojangOS() string {
	switch runtime.GOOS {
	case "darwin":
        return "osx"
    default:
        return runtime.GOOS
    }
}

func adoptiumOS() string {
	switch runtime.GOOS {
	case "darwin":
        return "mac"
    default:
        return runtime.GOOS
    }
}

func adoptiumArch() string {
	switch runtime.GOARCH {
	case "amd64":
        return "x64"
    case "arm64":
        return "aarch64"
    default:
        return runtime.GOARCH
    }
}

func isDirectoryEmpty(path string) bool {
	entries, _ := os.ReadDir(path)
	
	if len(entries) == 0 {
		return true
	}
	
	return false
}

func createInstanceDirs(dir string) error {
	dirs := []string{
		"mods",
        "saves",
        "config",
        "resourcepacks",
        "shaderpacks",
        "screenshots",
        "logs",
        "crash-reports",
    }

    for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			return err
        }
    }
    return nil
}

func extractTarGz(src, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
    }
	
    file, err := os.Open(src)
    if err != nil {
		return err
    }
    defer file.Close()

    gz, err := gzip.NewReader(file)
    if err != nil {
        return err
    }
    defer gz.Close()
	
    tr := tar.NewReader(gz)
	
    for {
		header, err := tr.Next()
        if err == io.EOF {
			break
        }
        if err != nil {
			return err
        }
		
        target := filepath.Join(dest, header.Name)
		
        switch header.Typeflag {
		case tar.TypeDir:
            os.MkdirAll(target, 0755)
        case tar.TypeReg:
            os.MkdirAll(filepath.Dir(target), 0755)
            f, err := os.Create(target)
            if err != nil {
				return err
            }
            io.Copy(f, tr)
			f.Chmod(fs.FileMode(header.Mode))
            f.Close()
        }
    }
    return nil
}

func loadCache() {
	_, err := os.Stat(internal.FileCache)

	if errors.Is(err, os.ErrNotExist) {
		cacheExists = false
		return
	}

	data, err := os.ReadFile(internal.FileCache)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(data, &hashedFiles)
}

func saveCache() {
	data, err := json.MarshalIndent(hashedFiles, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(internal.FileCache, data, 0644)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func checkCache(url, hash, dest string) (bool, error) {
	if !cacheExists || hash == "" {
		return false, nil
	}

	cachesMu.Lock()
	defer cachesMu.Unlock() 

	var foundIndex int = -1
	for i, j := range hashedFiles {
		if j.Hash == hash && j.URL == url {
			foundIndex = i
			break
		}
	}

	if foundIndex != -1 {
		file := hashedFiles[foundIndex]
		_, err := os.Stat(file.Path)
		if err == nil {
			freshHash, err := internal.FileHash(file.Path)
			if err == nil && freshHash != hash {
				fmt.Printf("Hash mismatch: expected %s got %s for %s\n", hash, freshHash, file.Path)
			}
			if err == nil && freshHash == file.Hash && freshHash == hash {
				//fmt.Printf("Using cached file from %s\n", file.Path)
				os.MkdirAll(filepath.Dir(dest), 0755)
				os.Remove(dest)
				return true, os.Symlink(file.Path, dest)
			}
		}
		hashedFiles = append(hashedFiles[:foundIndex], hashedFiles[foundIndex+1:]...)
	}
	// fmt.Printf("Cache miss: url=%s hash=%s\n", url, hash)
	return false, nil
}

func downloadFile(url string, dest string, hash string) error {
	var dl downloaded
	mu, _ := downloadMu.LoadOrStore(url, &sync.Mutex{})
    mu.(*sync.Mutex).Lock()
    defer mu.(*sync.Mutex).Unlock()

	cached, err := checkCache(url, hash, dest)
	if err != nil {
		return err
	}
	if cached {
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, 32*1024)
	// var downloaded int64
	total := resp.ContentLength

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := file.Write(buf[:n]); werr != nil {
				return werr
			}
			filename := filepath.Base(dest)
			if total > -1 {
				dl = downloaded{
					id: 0,
					finished: false,
				}

				if len(downloadedFiles) > 0 {
					dl.id = downloadedFiles[len(downloadedFiles) - 1].id + 1
				} else {
					dl.id = 1
				}
				downloadedFiles = append(downloadedFiles, &dl)

				pw := &progressWriter{
					dest:  dest,
					total: resp.ContentLength,
					file:  file,
					dled: &dl,
				}

				_, err = io.Copy(pw, resp.Body)
				if err != nil {
					return err
				}

			} else {
				fmt.Printf("\r%-40s     ", filename)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	hashed, err := internal.FileHash(dest)
	if err != nil {
		fmt.Printf("Error hashing downloaded file: %v\n", err)
	}

	cachesMu.Lock()
	hashedFiles = append(hashedFiles, internal.HashedFile{
		URL:  url,
		Hash: hashed,
		Path: dest,
	})
	cachesMu.Unlock()

	dl.finished = true
	return nil
}

func checkVersion(version string) {
    var fabricversion []internal.FabricVersion
    
    resp, err := rClient.R().
        SetResult(&fabricversion).
        Get(fmt.Sprintf(FM_VERSIONS_GAME, version))

    if err != nil || resp.IsError() {
        fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(fabricversion) == 0 {
        fmt.Printf("Error: Invalid version name. (%s)\n", version)
        os.Exit(1)
    }

}

func getVersionInfo(versionId string, rootdir string) internal.VersionInfo {
	checkVersion(versionId)
	var verManifest internal.VersionManifest

	resp, err := rClient.R().
		SetResult(&verManifest).
		Get(MM_V_MANIFEST)

	if err != nil || resp.IsError() {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var ver internal.VersionEntry
	
	for i, j := range verManifest.Versions {
		if j.ID == versionId {
			ver = verManifest.Versions[i];
			break
		}
	}

	if ver.ID == "" {
 		fmt.Println("Error: Version does not exist.")
		os.Exit(1)
	}

	var verInfo internal.VersionInfo
	
	downloadTo := filepath.Join(rootdir, "versions", versionId, versionId+".json")
	dir := filepath.Dir(downloadTo)
	os.MkdirAll(dir, os.ModePerm)
	err = downloadFile(ver.URL, downloadTo, "")

	if err != nil {
		fmt.Printf("Err: %v\n", err)
		os.Exit(1)
	}

	data, err := os.ReadFile(downloadTo)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(data, &verInfo)
	fmt.Printf("Got version info of %v\n", verInfo.ID)
	return verInfo
}

func downloadJar(verInfo *internal.VersionInfo, rootdir string) {
	downloadTo := filepath.Join(rootdir, "versions", verInfo.ID, verInfo.ID+".jar")
	dir := filepath.Dir(downloadTo)
	os.MkdirAll(dir, os.ModePerm)

	err := downloadFile(verInfo.Downloads.Client.URL, downloadTo, verInfo.Downloads.Client.Sha1)

	if err != nil {
		fmt.Printf("Err: %v\n", err)
		os.Exit(1)
	}
}

func downloadLibrary(artifact internal.Artifact, rootdir string) {
	libraryPath := filepath.Join(rootdir, "libraries", artifact.Path)
	err := downloadFile(artifact.URL, libraryPath, artifact.Sha1)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func downloadLibraries(libraries *[]internal.Library, rootdir string) {
	for _, j := range *libraries {
		if len(j.Rules) > 0 {
			for _, k := range j.Rules {
				if k.OS != nil && k.OS.Name == mojangOS() && k.Action == "allow" {
					downloadLibrary(j.Downloads.Artifact, rootdir)
				}
			}
			continue
		}

		downloadLibrary(j.Downloads.Artifact, rootdir)
	}
}

func downloadAsset(hash, rootdir string) {
	assetDir := filepath.Join(rootdir, "assets", "objects", hash[:2], hash)
	err := downloadFile(assetURL(hash), assetDir, hash)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func downloadAssets(s, id, rootdir string) {
    indexPath := filepath.Join(rootdir, "assets", "indexes", id+".json")
    err := downloadFile(s, indexPath, "")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    data, err := os.ReadFile(indexPath)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    var assetIndex internal.AssetIndex
    json.Unmarshal(data, &assetIndex)

    total := len(assetIndex.Objects)
    current := atomic.Int32{}

    sem := make(chan struct{}, 16)
    var wg sync.WaitGroup

    for _, i := range assetIndex.Objects {
        wg.Add(1)
        sem <- struct{}{}
        go func(hash string) {
            defer wg.Done()
            defer func() { <-sem }()
            downloadAsset(hash, rootdir)
            current.Add(1)
            fmt.Printf("\rDownloading assets: %d/%d", current.Load(), total)
        }(i.Hash)
    }
    wg.Wait()
    fmt.Println()
}

func downloadJRE(version int) {
    jrePath := filepath.Join(internal.JREPath, strconv.Itoa(version))
    
    if _, err := os.Stat(jrePath); err == nil {
        //fmt.Printf("JRE %d already exists, skipping.\n", version)
        return
    }

    tmpDir, err := os.MkdirTemp("", "anvil-java-*")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer os.RemoveAll(tmpDir)

    extension := ".tar.gz"
    if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
        extension = ".zip"
    }

    downloadto := filepath.Join(tmpDir, "jre"+extension)
    err = downloadFile(fmt.Sprintf(A_JAVA, version), downloadto, "")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    if err := extractTarGz(downloadto, jrePath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func mavenPath(name string) string {
    // "net.fabricmc:fabric-loader:0.15.0" -> "net/fabricmc/fabric-loader/0.15.0/fabric-loader-0.15.0.jar"
    parts := strings.Split(name, ":")
    group := strings.ReplaceAll(parts[0], ".", "/")
    artifact := parts[1]
    version := parts[2]
    return fmt.Sprintf("%s/%s/%s/%s-%s.jar", group, artifact, version, artifact, version)
}

func mavenURL(baseURL, name string) string {
    return baseURL + mavenPath(name)
}

func getFabricProfile(version string) internal.FabricProfile {
	var fabricversion []internal.FabricLoaderVersion
	var fabricProfile internal.FabricProfile
    
    resp, err := rClient.R().
        SetResult(&fabricversion).
        Get(fmt.Sprintf(FM_LOADER, version))

    if err != nil || resp.IsError() {
        fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err = rClient.R().
        SetResult(&fabricProfile).
        Get(fmt.Sprintf(FM_PROFILE, version, fabricversion[0].Loader.Version))

    if err != nil || resp.IsError() {
        fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fabricProfile.LoaderVersion = fabricversion[0].Loader.Version
	return fabricProfile
}

func downloadFabricLibrary(lib internal.FabricLibrary, rootdir string) {
	path := filepath.Join(rootdir, "libraries", mavenPath(lib.Name))
	url := mavenURL(lib.URL, lib.Name)

	err := downloadFile(url, path, "")

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func downloadFabricLibraries(fabricLibraries []internal.FabricLibrary, rootdir string) {
	for _, j := range fabricLibraries {
		downloadFabricLibrary(j, rootdir)
	}
}

func New(args ...string) {
	// Setup Minecraft and Fabric
	
	name := args[0]
	version := args[1]
	dir := filepath.Join(internal.InstancesPath, name)
	rClient = internal.NewClient()

	err := os.MkdirAll(dir, os.ModePerm)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !isDirectoryEmpty(dir) {
		fmt.Printf("Error: Directory (%s) is not empty.\n", dir)
		os.Exit(1)
	}

	if err := createInstanceDirs(dir); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	verInfo := getVersionInfo(version, dir)

	loadCache()
	
	downloadJar(&verInfo, dir)
	downloadJRE(verInfo.JavaVersion.MajorVersion)
	downloadLibraries(&verInfo.Libraries, dir)
	downloadAssets(verInfo.AssetIndex.URL, verInfo.AssetIndex.ID, dir)

	// Setting up minecraft finished.
	// Now its time to set up Fabric.

	fabricProfile := getFabricProfile(version)

	downloadFabricLibraries(fabricProfile.Libraries, dir)

	// Installed Fabric.
	// Adding an anvil.json so we know its anvil instance.

	instanceInfo := internal.Instance{
		Name:    name,
		Version: version,
		LoaderVersion: fabricProfile.LoaderVersion,
		MainClass: fabricProfile.MainClass,
		FabricProfile: fabricProfile,
		Mods:    []internal.Mod{},
	}

	data, err := json.MarshalIndent(instanceInfo, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(filepath.Join(dir, "anvil.json"), data, 0644)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	saveCache()
	Select(name)
}
