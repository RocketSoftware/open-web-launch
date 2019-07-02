package jnlp

import (
	"io/ioutil"
	"log"
	"sync"

	"github.com/pkg/errors"
	"github.com/rocketsoftware/open-web-launch/gui"
	launcher_utils "github.com/rocketsoftware/open-web-launch/launcher/utils"
	"github.com/rocketsoftware/open-web-launch/utils/download"
)

func (launcher *Launcher) UninstallByFilename(filename string, showGUI bool) error {
	log.Printf("uninstall using filename %s", filename)
	if showGUI {
		launcher.gui = gui.NewNativeGUI()
	}
	var wg sync.WaitGroup
	var err error
	wg.Add(1)
	go func() {
		defer func() {
			if err == nil {
				launcher.gui.Terminate()
			}
			wg.Done()
		}()
		launcher.gui.WaitForWindow()
		var filedata []byte
		if filedata, err = ioutil.ReadFile(filename); err != nil {
			return
		}
		if err = launcher.uninstallUsingFileData(filedata); err != nil {
			return
		}
	}()
	if err = launcher.gui.Start(launcher.WindowTitle); err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func (launcher *Launcher) UninstallByURL(url string, showGUI bool) error {
	log.Printf("uninstall using URL %s", url)
	var err error
	url = launcher.normalizeURL(url)
	if showGUI {
		launcher.gui = gui.NewNativeGUI()
	}
	launcher.gui = gui.NewNativeGUI()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			if err == nil {
				launcher.gui.Terminate()
			}
			wg.Done()
		}()
		launcher.gui.WaitForWindow()
		var filedata []byte
		if filedata, err = download.ToMemory(url); err != nil {
			return
		}
		if err = launcher.uninstallUsingFileData(filedata); err != nil {
			return
		}
	}()
	if err = launcher.gui.Start(launcher.WindowTitle); err != nil {
		return err
	}
	wg.Wait()
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
