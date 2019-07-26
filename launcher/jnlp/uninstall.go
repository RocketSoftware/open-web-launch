package jnlp

import (
	"io/ioutil"
	"sync"

	"github.com/pkg/errors"
	"github.com/rocketsoftware/open-web-launch/gui"
	launcher_utils "github.com/rocketsoftware/open-web-launch/launcher/utils"
	"github.com/rocketsoftware/open-web-launch/utils/download"
	"github.com/rocketsoftware/open-web-launch/utils/log"
)

func (launcher *Launcher) UninstallByFilename(filename string, showGUI bool) error {
	return launcher.uninstallByFilenameOrURL(filename, showGUI, false)
}

func (launcher *Launcher) UninstallByURL(url string, showGUI bool) error {
	return launcher.uninstallByFilenameOrURL(url, showGUI, true)
}

func (launcher *Launcher) uninstallByFilenameOrURL(filenameOrURL string, showGUI bool, isURL bool) error {
	log.Printf("uninstall using %s", filenameOrURL)
	if showGUI {
		launcher.gui = gui.New()
		launcher.gui.SetLogFile(launcher.logFile)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var err error
		defer func() {
			if err != nil {
				log.Println(err)
			}
			wg.Done()
		}()
		launcher.gui.WaitForWindow()
		var filedata []byte
		if isURL {
			url := launcher.normalizeURL(filenameOrURL)
			filedata, err = download.ToMemory(url)
		} else {
			filedata, err = ioutil.ReadFile(filenameOrURL)
		}
		if err != nil {
			launcher.gui.SendErrorMessage(err)
			return
		}
		if err = launcher.uninstallUsingFileData(filedata); err != nil {
			launcher.gui.SendErrorMessage(err)
			return
		}
	}()
	if err := launcher.gui.Start(launcher.WindowTitle); err != nil {
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
	launcher.gui.SetTitle(jnlpFile.Title())
	launcher.gui.SetProgressMax(3)
	launcher.removeShortcuts(jnlpFile)
	launcher.gui.ProgressStep()
	launcher_utils.RemoveResourceDir(launcher.WorkDir, filedata)
	launcher.gui.ProgressStep()
	launcher.uninstallApp(jnlpFile)
	launcher.gui.ProgressStep()
	launcher.gui.SendTextMessage("Uninstall complete")
	return nil
}
