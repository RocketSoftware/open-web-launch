package java

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/rocketsoftware/open-web-launch/utils"

	"github.com/pkg/errors"
)

var javaExecutable string
var javaSource string
var jarSignerExecutable string

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

// Version returns detailed Java version information, e.g.
// java version "1.8.0_171" Java(TM) SE Runtime Environment (build 1.8.0_171-b11) Java HotSpot(TM) 64-Bit Server VM (build 25.171-b11, mixed mode)
func Version() (string, error) {
	cmd := exec.Command(javaExecutable, "-version")
	utils.HideWindow(cmd)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "Unable to obtain Java version")
	}
	re := regexp.MustCompile("[^!-~\t ]")
	output := re.ReplaceAllLiteralString(string(outputBytes), " ")
	return output, nil
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

func init() {
	javaExecutable = getJavaExecutable()
	jarSignerExecutable = getJARSignerExecutable()
}
