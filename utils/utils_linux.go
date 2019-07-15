package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/image/font"
)

func HideWindow(cmd *exec.Cmd) {
}

func BreakAwayFromParent(cmd *exec.Cmd) {
	var sysProcAttr *syscall.SysProcAttr
	if cmd.SysProcAttr != nil {
		sysProcAttr = cmd.SysProcAttr
	} else {
		sysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr = sysProcAttr
	}
	sysProcAttr.Setsid = true
}

func HideJavaWindowIfNeeded(cmd *exec.Cmd) {
}

func LoadIconAndSetForWindow(windowTitle string) error {
	return nil
}

func LoadFont(fontFace string, size int, scaling float64) (font.Face, error) {
	return nil, errors.Errorf("LoadFont not implemented for platform %s", runtime.GOOS)
}

func CallLibrary(path string, funcName string, arg string) (err error) {
	return nil
}

func CreateDesktopShortcut(src, title, description, iconSrc string, arguments ...string) error {
	return nil
}

func CreateStartMenuShortcut(src, folder, title, description, iconSrc string, arguments ...string) error {
	return nil
}

func RemoveDesktopShortcut(title string) error {
	return nil
}

func RemoveStartMenuFolder(folder string) error {
	return nil
}

func ShowUsage(productTitle, productVersion, text string) {
	fmt.Fprintf(os.Stderr, text)
}

func InstallApp(app *AppInfo) error {
	return nil
}

func UninstallApp(title string) error {
	return nil
}

func OpenTextFile(filename string) error {
	cmd := exec.Command("xdg-open", filename)
	return cmd.Start()
}
