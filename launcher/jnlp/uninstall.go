package jnlp

import (
	"io/ioutil"
	"log"

	launcher_utils "github.com/rocketsoftware/open-web-launch/launcher/utils"
	"github.com/rocketsoftware/open-web-launch/utils/download"
	"github.com/pkg/errors"
)

func (launcher *Launcher) UninstallByFilename(filename string) error {
	log.Printf("uninstall using filename %s", filename)
	var err error
	var filedata []byte
	if filedata, err = ioutil.ReadFile(filename); err != nil {
		return err
	}
	if err = launcher.uninstallUsingFileData(filedata); err != nil {
		return err
	}
	return nil
}

func (launcher *Launcher) UninstallByURL(url string) error {
	log.Printf("uninstall using URL %s", url)
	var err error
	url = launcher.normalizeURL(url)
	var filedata []byte
	if filedata, err = download.ToMemory(url); err != nil {
		return err
	}
	if err = launcher.uninstallUsingFileData(filedata); err != nil {
		return err
	}
	return nil
}

func (launcher *Launcher) uninstallUsingFileData(filedata []byte) error {
	var err error
	var jnlpFile *JNLP
	if jnlpFile, err = Decode(filedata); err != nil {
		return errors.Wrap(err, "parsing JNLP")
	}
	launcher.removeShortcuts(jnlpFile)
	launcher_utils.RemoveResourceDir(launcher.WorkDir, filedata)
	return nil
}
