package verifier

import (
	"archive/zip"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/rocketsoftware/open-web-launch/java"
	"github.com/rocketsoftware/open-web-launch/utils"
	"github.com/pkg/errors"
)

const (
	verboseMessage = "Re-run jarsigner with the -verbose option for more details."
)

func VerifyWithJARSigner(jar string, verbose bool) error {
	jarSignerExecutable := java.JARSigner()
	if jarSignerExecutable == "" {
		return nil
	}
	var cmd *exec.Cmd
	if verbose {
		cmd = exec.Command(jarSignerExecutable, "-verify", "-verbose", jar)
	} else {
		cmd = exec.Command(jarSignerExecutable, "-verify", jar)
	}
	utils.HideWindow(cmd)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "unable to run jarsigner")
	}
	output := string(stdoutStderr)
	if !strings.Contains(output, "jar verified.") {
		if !verbose && strings.Contains(output, verboseMessage) {
			output = strings.Replace(output, verboseMessage, " See log file for more details", 1)
			err := VerifyWithJARSigner(jar, true)
			log.Println(err)
		}
		return errors.New(output)
	}
	return nil
}

func GetJARCertificate(jar string) ([]byte, error) {
	reader, err := zip.OpenReader(jar)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var certificate []byte
	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "META-INF/") {
			dir, filename := path.Split(file.Name)
			if dir != "META-INF/" {
				continue
			}
			ext := path.Ext(filename)
			if ext == ".RSA" || ext == ".DSA" {
				data, err := getFileContent(file)
				if err != nil {
					return nil, errors.Wrapf(err, "unable to read certificate file %s", filename)
				}
				certificate = data
			}
		}
	}
	if certificate == nil {
		return nil, errors.Errorf("unable to find certificate in JAR file %s", filepath.Base(jar))
	}
	return certificate, nil
}

func getFileContent(file *zip.File) ([]byte, error) {
	fileReader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer fileReader.Close()
	return ioutil.ReadAll(fileReader)
}
