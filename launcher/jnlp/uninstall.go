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
	log.Printf("uninstall using filename %s", filename)
	if showGUI {
		launcher.gui = gui.New()
		launcher.gui.SetLogFile(launcher.logFile)
	}
	var wg sync.WaitGroup
	var err error
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		launcher.gui.WaitForWindow()
		var filedata []byte
		if filedata, err = ioutil.ReadFile(filename); err != nil {
			launcher.gui.SendErrorMessage(err)
			return
		}
		if err = launcher.uninstallUsingFileData(filedata); err != nil {
			launcher.gui.SendErrorMessage(err)
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
		launcher.gui = gui.New()
		launcher.gui.SetLogFile(launcher.logFile)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		launcher.gui.WaitForWindow()
		var filedata []byte
		if filedata, err = download.ToMemory(url); err != nil {
			launcher.gui.SendErrorMessage(err)
			return
		}
		if err = launcher.uninstallUsingFileData(filedata); err != nil {
			launcher.gui.SendErrorMessage(err)
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
