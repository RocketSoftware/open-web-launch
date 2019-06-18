package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rocketsoftware/open-web-launch/utils"
	"github.com/pkg/errors"
)

// GenerateResourcesDirName generates a directory name for resource files.
func GenerateResourcesDirName(workDir string, filedata []byte) string {
	hasher := sha256.New()
	hasher.Write(filedata)
	dir := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil)[:16])
	return filepath.Join(workDir, dir)
}

// RemoveResourceDir removes resource directory for an app
// filedata - content of JNLP file for app
func RemoveResourceDir(workDir string, filedata []byte) error {
	resourceDir := GenerateResourcesDirName(workDir, filedata)
	log.Printf("removing resource directory %s", resourceDir)
	if err := os.RemoveAll(resourceDir); err != nil {
		log.Printf("warning: removing resource directory: %v", err)
	}
	return nil
}

// ParseCodebaseURL parses codebase URL, adds trailing slash if needed
func ParseCodebaseURL(codebase string) (*url.URL, error) {
	if !strings.HasSuffix(codebase, "/") {
		codebase += "/"
	}
	codebaseURL, err := url.Parse(codebase)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse codebase URL")
	}
	return codebaseURL, nil
}

// ArchSynonyms maps JWS architectures to Golang ones
var ArchSynonyms = map[string]string{
	"x86":    "386",
	"i386":   "386",
	"i686":   "386",
	"x64":    "amd64",
	"x86_64": "amd64",
}

// OSSynonyms maps JWS OSes to Golang ones
var OSSynonyms = map[string]string{
	"mac os":       "darwin",
	"mac os x":     "darwin",
	"windows 95":   "windows",
	"windows 98":   "windows",
	"windows nt":   "windows",
	"windows 2000": "windows",
	"windows 2003": "windows",
}

func AreResourcesRelevantForCurrentPlatform(os string, arch string) bool {
	if os == "" {
		return true
	}
	oses := utils.SplitEscapedString(os)
	osMatches := false
	for _, os := range oses {
		osInLowercase := strings.ToLower(os)
		if osInLowercase == runtime.GOOS || OSSynonyms[osInLowercase] == runtime.GOOS {
			osMatches = true
			break
		}
	}
	if !osMatches {
		return false
	}
	if arch == "" {
		return true
	}
	arches := utils.SplitEscapedString(arch)
	for _, arch := range arches {
		archInLowercase := strings.ToLower(arch)
		if archInLowercase == runtime.GOARCH || ArchSynonyms[archInLowercase] == runtime.GOARCH {
			return true
		}
	}
	return false
}
