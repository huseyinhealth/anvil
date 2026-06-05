package internal

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
)

func FileHash(path string) (string, error) {
    file, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer file.Close()

    h := sha1.New()
    if _, err := io.Copy(h, file); err != nil {
        return "", err
    }

    return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func RemoveFromSlice[T comparable](slice []T, target T) []T {
    for i, v := range slice {
        if v == target { // comparable sayesinde derleyici buraya kızmaz!
            return append(slice[:i], slice[i+1:]...)
        }
    }
    return slice // Eleman bulunamazsa orijinal slice'ı dön
}

func GetSelectedInstance() (string, int) {
    buff, err := os.ReadFile(filepath.Join(AnvilHome, ".selected"))

    if err != nil {
        // fmt.Printf("Error: %v\n", err)
        // os.Exit(1)
        return "Couldn't get selected instance.", 1
    }
    return string(buff), 0
}

func NewClient() *resty.Client {
    return resty.New().SetHeader("User-Agent", "huseyinhealth/anvil/1.0.0 (hello@cnhsync.dev)")
}
