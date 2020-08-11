package utils

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"github.com/pkg/errors"
)

func HideWindow(cmd *exec.Cmd) {
	var sysProcAttr *syscall.SysProcAttr
	if cmd.SysProcAttr != nil {
		sysProcAttr = cmd.SysProcAttr
	} else {
		sysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr = sysProcAttr
	}
	sysProcAttr.HideWindow = true
}

const (
	CREATE_BREAKAWAY_FROM_JOB = 0x01000000
)

func BreakAwayFromParent(cmd *exec.Cmd) {
	var sysProcAttr *syscall.SysProcAttr
	if cmd.SysProcAttr != nil {
		sysProcAttr = cmd.SysProcAttr
	} else {
		sysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr = sysProcAttr
	}
	sysProcAttr.CreationFlags |= CREATE_BREAKAWAY_FROM_JOB
}

func HideJavaWindowIfNeeded(cmd *exec.Cmd) {
	if strings.HasSuffix(cmd.Path, "javaw.exe") {
		HideWindow(cmd)
	}
}

func LoadIconAndSetForWindow(windowTitle string) error {
	hwnd, err := findWindowByName(windowTitle)
	if err != nil {
		return err
	}
	instance, err := getModuleHandle()
	if err != nil {
		return err
	}
	icon, err := loadIconResource(instance, cIDC_ICON)
	if err != nil {
		return err
	}
	return setWindowIcon(hwnd, icon)
}

var (
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	pGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

func getModuleHandle() (syscall.Handle, error) {
	ret, _, err := pGetModuleHandleW.Call(uintptr(0))
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

var (
	user32 = syscall.NewLazyDLL("user32.dll")

	pLoadIconW   = user32.NewProc("LoadIconW")
	pFindWindowW = user32.NewProc("FindWindowW")
)

const (
	cIDC_ICON     = 1
	cIDC_ARROW    = 32512
	cIDI_QUESTION = 32514
)

func findWindowByName(windowName string) (syscall.Handle, error) {
	ret, _, err := pFindWindowW.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowName))),
	)
	if ret == 0 {
		return 0, errors.Wrap(err, "unable to find window")
	}
	return syscall.Handle(ret), nil
}

func loadIconResource(instance syscall.Handle, iconName uint32) (syscall.Handle, error) {
	ret, _, err := pLoadIconW.Call(
		uintptr(instance),
		uintptr(uint16(iconName)),
	)
	if ret == 0 {
		return 0, errors.Wrapf(err, "unable to load icon resource with id %d", iconName)
	}
	return syscall.Handle(ret), nil
}

func setClassLongPtr(hwnd syscall.Handle, index int, value uintptr) error {
	name := "SetClassLongPtrW"
	if runtime.GOARCH == "386" {
		name = "SetClassLongW"
	}
	pSetClassLongPtrW := user32.NewProc(name)
	ret, _, err := pSetClassLongPtrW.Call(
		uintptr(hwnd),
		uintptr(index),
		value)
	if ret == 0 {
		return errors.Wrap(err, "unable to set class long ptr")
	}
	return nil
}

const (
	cGCL_HICON   = -14
	cGCL_HICONSM = -34
)

func setWindowIcon(hwnd syscall.Handle, icon syscall.Handle) error {
	err := setClassLongPtr(hwnd, cGCL_HICON, uintptr(icon))
	if err != nil {
		return errors.Wrap(err, "unable to set icon for window")
	}
	return nil
}

func LoadFont(fontFace string, size int, scaling float64) (font.Face, error) {
	windowsDir := os.Getenv("WINDIR")
	fontFile := filepath.Join(windowsDir, "Fonts", fontFace+".ttf")
	fontData, err := ioutil.ReadFile(fontFile)
	if err != nil {
		return nil, err
	}
	trueTypeFont, err := truetype.Parse(fontData)
	if err != nil {
		return nil, err
	}
	return truetype.NewFace(trueTypeFont, &truetype.Options{Size: scaling * float64(size), Hinting: font.HintingNone, DPI: 96}), nil
}

func CallLibrary(path string, funcName string, arg string) (err error) {
	var lib syscall.Handle
	var proc uintptr
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "unable to call %s from %s library with arg %s", funcName, path, arg)
		}
	}()
	lib, err = syscall.LoadLibrary(path)
	if err != nil {
		return
	}
	proc, err = syscall.GetProcAddress(lib, funcName)
	if err != nil {
		return
	}
	byteArg := make([]byte, len(arg)+10)
	copy(byteArg, []byte(arg))
	var nargs uintptr = 1
	syscall.Syscall(uintptr(proc), nargs, uintptr(unsafe.Pointer(&byteArg[0])), 0, 0)
	return nil
}

const wshShellClass = "WScript.Shell"

func CreateDesktopShortcut(src, title, description, iconSrc string, arguments ...string) error {
	var err error
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "unable to create a desktop shortcut for %v with title %s", arguments, title)
		}
	}()
	desktopFolder, err := GetDesktopFolder()
	if err != nil {
		return err
	}
	err = CreateShortcut(src, desktopFolder, title, description, iconSrc, arguments...)
	return nil
}

func CreateStartMenuShortcut(src, folder, title, description, iconSrc string, arguments ...string) error {
	var err error
	var programsDir string
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "unable to create a Start Menu shortcut for %v with title %s", arguments, title)
		}
	}()
	programsDir, err = getProgramsStartMenuDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(programsDir, folder)
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			err = errors.Wrapf(err, "unable to create folder in Start Menu'%s'", dir)
			return err
		}
	}
	err = CreateShortcut(src, dir, title, description, iconSrc, arguments...)
	return nil
}

func CreateShortcut(src, dstFolder, title, description, iconSrc string, arguments ...string) error {
	var err error
	if err = ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY); err != nil {
		return errors.Wrap(err, "unable to initialize COM library")
	}
	defer ole.CoUninitialize()
	wshShellObject, err := oleutil.CreateObject(wshShellClass)
	if err != nil {
		return errors.Wrap(err, "unable to create COM object "+wshShellClass)
	}
	defer wshShellObject.Release()
	wshShell, err := wshShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer wshShell.Release()
	link := filepath.Join(dstFolder, title+".lnk")
	if _, err := os.Stat(link); err == nil {
		if err := os.Remove(link); err != nil {
			return errors.Wrap(err, "unable to delete already existing shortcut")
		}
	}
	createShortcutResult, err := oleutil.CallMethod(wshShell, "CreateShortcut", link)
	if err != nil {
		return err
	}
	shortcut := createShortcutResult.ToIDispatch()
	_, err = oleutil.PutProperty(shortcut, "TargetPath", src)
	if err != nil {
		return errors.Wrap(err, "unable to set TargetPath for the shortcut")
	}
	if iconSrc != "" {
		_, err = oleutil.PutProperty(shortcut, "IconLocation ", iconSrc)
		if err != nil {
			return errors.Wrap(err, "unable to set icon for the shortcut")
		}
	}
	if description != "" {
		_, err = oleutil.PutProperty(shortcut, "Description", description)
		if err != nil {
			return errors.Wrap(err, "unable to set description for the shortcut")
		}
	}
	_, err = oleutil.PutProperty(shortcut, "Arguments", prepareShortcutArguments(arguments...))
	if err != nil {
		return errors.Wrap(err, "unable to set arguments for the shortcut")
	}
	_, err = oleutil.CallMethod(shortcut, "Save")
	if err != nil {
		return errors.Wrap(err, "unable to save the shortcut")
	}
	return nil
}

func prepareShortcutArguments(arguments ...string) string {
	for i, arg := range arguments {
		if strings.Contains(arg, " ") {
			arguments[i] = QuoteString(arg)
		}
	}
	return strings.Join(arguments, " ")
}

func RemoveDesktopShortcut(title string) error {
	var err error
	var desktopFolder string
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "unable to remove desktop shortcut %s ", title)
		}
	}()
	desktopFolder, err = GetDesktopFolder()
	if err != nil {
		return err
	}
	link := filepath.Join(desktopFolder, title+".lnk")
	if _, err = os.Stat(link); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(link)
}

func RemoveStartMenuFolder(folder string) error {
	var err error
	var programsDir string
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "unable to remove Start Menu folder %s", folder)
		}
	}()
	programsDir, err = getProgramsStartMenuDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(programsDir, folder)
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dir)
}

func getProgramsStartMenuDir() (string, error) {
	startMenuFolder, err := GetStartMenuFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(startMenuFolder, "Programs"), nil
}

func GetDesktopFolder() (string, error) {
	return windows.KnownFolderPath(windows.FOLDERID_Desktop, 0)
}

func GetStartMenuFolder() (string, error) {
	return windows.KnownFolderPath(windows.FOLDERID_StartMenu, 0)
}

var (
	pMessageBoxW = user32.NewProc("MessageBoxW")
)

const (
	cMB_ICONINERROR     = 0x00000010
	cMB_ICONINFORMATION = 0x00000040
)

func ShowUsage(productTitle, productVersion, text string) {
	caption := productTitle + " " + productVersion
	pMessageBoxW.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(caption))),
		uintptr(cMB_ICONINFORMATION),
	)
}

func ShowFatalError(text string) {
	caption := "Error"
	pMessageBoxW.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(caption))),
		uintptr(cMB_ICONINERROR),
	)
}

func InstallApp(app *AppInfo) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall\`+app.Title, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	if err := key.SetStringValue("DisplayName", app.Title); err != nil {
		return err
	}
	if err := key.SetStringValue("UninstallString", app.UninstallString); err != nil {
		return err
	}
	if err := key.SetStringValue("DisplayIcon", app.Icon); err != nil {
		return err
	}
	if err := key.SetStringValue("Version", app.Version); err != nil {
		return err
	}
	if err := key.SetStringValue("URLInfoAbout", app.URL); err != nil {
		return err
	}
	if err := key.SetStringValue("Publisher", app.Publisher); err != nil {
		return err
	}
	if err := key.SetDWordValue("NoModify", 1); err != nil {
		return err
	}
	if err := key.SetDWordValue("NoRepair", 1); err != nil {
		return err
	}
	return nil
}

func UninstallApp(title string) error {
	if err := registry.DeleteKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall\`+title); err != nil {
		return err
	}
	return nil
}

func OpenTextFile(filename string) error {
	cmd := exec.Command("notepad.exe", filename)
	return cmd.Start()
}
