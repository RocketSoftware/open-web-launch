package settings

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/rocketsoftware/open-web-launch/utils"

	"github.com/pkg/errors"
)

var javaExecutable string
var javaSource string
var jarSignerExecutable string
var disableVerification bool
var addAppToControlPanel bool
var currentJavaVersion *JavaVersion

type Version struct {
	Major       int
	Minor       int
	AllowHigher bool
}

func EnsureJavaExecutableAvailability() error {
	if filepath.IsAbs(javaExecutable) {
		if _, err := os.Stat(javaExecutable); err != nil {
			return errors.Errorf("Java location configured using %s but Java executable %s is missing", JavaSource(), javaExecutable)
		}
		log.Printf("java executable is %s found using %s\n", javaExecutable, JavaSource())
		return nil
	}
	fullPath, err := exec.LookPath(javaExecutable)
	if err != nil {
		return errors.Errorf("java executable %s wasn't found in PATH", javaExecutable)
	}
	log.Printf("java executable is found in PATH: %s\n", fullPath)
	javaExecutable = fullPath
	return nil
}

func EnsureJARSignerAvailability() error {
	if filepath.IsAbs(jarSignerExecutable) {
		if _, err := os.Stat(jarSignerExecutable); err != nil {
			errMessage := fmt.Sprintf("jarsigner executable %s wasn't found", jarSignerExecutable)
			jarSignerExecutable = ""
			return errors.New(errMessage)
		}
		log.Printf("jarsigner executable is %s\n", jarSignerExecutable)
		return nil
	}
	fullPath, err := exec.LookPath(jarSignerExecutable)
	if err != nil {
		errMessage := fmt.Sprintf("jarsigner %s wasn't found in PATH", jarSignerExecutable)
		jarSignerExecutable = ""
		return errors.New(errMessage)
	}
	log.Printf("jarsigner executable is found in PATH: %s", fullPath)
	jarSignerExecutable = fullPath
	return nil
}

func Java() string {
	return javaExecutable
}

func JARSigner() string {
	return jarSignerExecutable
}

func JavaSource() string {
	return javaSource
}

func IsVerificationDisabled() bool {
	return disableVerification
}

func AddAppToControlPanel() bool {
	return addAppToControlPanel
}

// GetJavaVersionString returns detailed Java version information, e.g.
// java version "1.8.0_171" Java(TM) SE Runtime Environment (build 1.8.0_171-b11) Java HotSpot(TM) 64-Bit Server VM (build 25.171-b11, mixed mode)
func GetJavaVersionString() (string, error) {
	cmd := exec.Command(javaExecutable, "-version")
	utils.HideWindow(cmd)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "unable to obtain Java version")
	}
	re := regexp.MustCompile("[^!-~\t ]")
	output := re.ReplaceAllLiteralString(string(outputBytes), " ")
	return output, nil
}

type JavaVersion struct {
	Major       int
	Minor       int
	AllowHigher bool
}

// GetJavaVersion returns major ans minor Java version
func GetJavaVersion() (javaVersion *JavaVersion, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "unable to detect Java version")
		}
	}()
	if currentJavaVersion != nil {
		return currentJavaVersion, nil
	}
	versionOutput, err := GetJavaVersionString()
	if err != nil {
		return
	}
	firstQuoteIndex := strings.Index(versionOutput, `"`)
	if firstQuoteIndex == -1 {
		err = errors.Wrapf(err, `unable to locate Java version: double quote not found: "%s"`, versionOutput)
		return
	}
	secondQuoteIndex := strings.Index(versionOutput[firstQuoteIndex+1:], `"`)
	if secondQuoteIndex == -1 {
		err = errors.Wrapf(err, `unable to locate Java version: second double quote not found: "%s"`, versionOutput)
		return
	}
	version := versionOutput[firstQuoteIndex+1 : secondQuoteIndex+firstQuoteIndex+1]
	javaVersion, err = ParseJavaVersion(version)
	if err != nil {
		return
	}
	currentJavaVersion = javaVersion
	return
}

func ParseJavaVersion(version string) (*JavaVersion, error) {
	allowHigher := false
	ver := version
	if strings.HasSuffix(version, "+") {
		ver = strings.TrimSuffix(version, "+")
		allowHigher = true
	}
	parts := strings.Split(ver, ".")
	if len(parts) == 0 {
		return nil, errors.Errorf(`unable to parse Java version "%s"`, version)
	}
	var major, minor int64
	var err error
	if major, err = strconv.ParseInt(parts[0], 10, 8); err != nil {
		return nil, errors.Wrapf(err, `unable to parse major version "%s"`, parts[0])
	}
	if len(parts) > 1 {
		if minor, err = strconv.ParseInt(parts[1], 10, 8); err != nil {
			return nil, errors.Wrapf(err, `unable to parse minor version "%s"`, parts[1])
		}
	}
	return &JavaVersion{int(major), int(minor), allowHigher}, nil
}

func CurrentJavaVersionMatches(version *JavaVersion) bool {
	if currentJavaVersion == nil {
		return false
	}
	if currentJavaVersion.Major < version.Major {
		return false
	}
	if currentJavaVersion.Major > version.Major && version.AllowHigher {
		return true
	}
	if currentJavaVersion.Major == version.Major {
		if currentJavaVersion.Minor < version.Minor {
			return false
		}
		if currentJavaVersion.Minor == version.Minor || (currentJavaVersion.Minor > version.Minor && version.AllowHigher) {
			return true
		}
	}
	return false
}

func getJavaExecutableUsingJavaHome(showConsole bool) (string, error) {
	javaSource = "JAVA_HOME environment variable - " + os.Getenv("JAVA_HOME")
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		return "", errors.New("JAVA_HOME environment variable is not set")
	}
	java := "java"
	if runtime.GOOS == "windows" {
		if showConsole {
			java = "java.exe"
		} else {
			java = "javaw.exe"
		}
	}
	return filepath.Join(javaHome, "bin", java), nil
}

// UseJavaDir forces to use Java installation from directory dir.
// Returns absolute path to the specified directory.
func UseJavaDir(dir string) (string, error) {
	javaSource = `-javadir '` + dir + `' command line argument`
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return "", errors.Wrapf(err, `invalid javadir '%s'`, dir)
	}
	fileInfo, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return "", errors.Errorf(`javadir '%s' doesn't exist`, dir)
	}
	if !fileInfo.IsDir() {
		return "", errors.Errorf(`javadir '%s' is not a directory`, dir)
	}
	javaExecutable = getJavaExecutableUsingJavaDir(absPath)
	jarSignerExecutable = getJARSignerExecutableUsingJavaDir(absPath)
	return absPath, nil
}

func ShowConsole() {
	if runtime.GOOS == "windows" && javaExecutable != "" {
		javaDir := filepath.Dir(javaExecutable)
		javaExecutable = filepath.Join(javaDir, "java.exe")
	}
}

func init() {
	javaExecutable = getJavaExecutable()
	jarSignerExecutable = getJARSignerExecutable()
	disableVerification = getDisableVerificationSetting()
	addAppToControlPanel = getAddAppToControlPanelSetting()
}
