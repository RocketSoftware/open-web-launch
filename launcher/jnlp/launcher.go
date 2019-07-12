package jnlp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rocketsoftware/open-web-launch/launcher"
	launcher_utils "github.com/rocketsoftware/open-web-launch/launcher/utils"

	"github.com/pkg/errors"
	"github.com/rocketsoftware/open-web-launch/gui"
	"github.com/rocketsoftware/open-web-launch/settings"
	"github.com/rocketsoftware/open-web-launch/utils"
	"github.com/rocketsoftware/open-web-launch/utils/download"
	"github.com/rocketsoftware/open-web-launch/verifier"
)

var errCancelled = errors.New("cancelled by user")

// Launcher is a JNLP Launcher
type Launcher struct {
	WorkDir           string // Working directory
	WindowTitle       string // Title of GUI Window
	jnlp              *JNLP
	jnlpOld           *JNLP
	filedata          []byte
	resourceDir       string
	relevantResources []*Resources
	codebaseURL       *url.URL
	cmd               *exec.Cmd
	gui               *gui.GUI
	options           *launcher.Options
	cert              []byte
}

// New creates a new JNLP Launcher
func NewLauncher() *Launcher {
	return &Launcher{
		WorkDir: ".",
	}
}

func (launcher *Launcher) SetWindowTitle(title string) {
	launcher.WindowTitle = title
}

func (launcher *Launcher) SetWorkDir(dir string) {
	launcher.WorkDir = dir
}

// RunByURL runs a JNLP file by URL
func (launcher *Launcher) RunByURL(url string) error {
	var err error
	log.Printf("Processing %s\n", url)
	url = launcher.normalizeURL(url)
	launcher.gui = gui.New()
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
			launcher.gui.SendErrorMessage(err)
			return
		}
		if err = launcher.run(filedata); err != nil {
			launcher.gui.SendErrorMessage(err)
			return
		}
	}()
	if err = launcher.gui.Start(launcher.WindowTitle); err != nil {
		return err
	}
	wg.Wait()
	return err
}

func (launcher *Launcher) SetOptions(options *launcher.Options) {
	launcher.options = options
}

// RunByFilename runs a JNLP file
func (launcher *Launcher) RunByFilename(filename string) error {
	var err error
	log.Printf("Processing %s\n", filename)
	launcher.gui = gui.New()
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
		if filedata, err = ioutil.ReadFile(filename); err != nil {
			launcher.gui.SendErrorMessage(err)
			return
		}
		filedata, err = launcher.checkForUpdate(filedata)
		if err != nil {
			launcher.gui.SendErrorMessage(err)
			return
		}
		if err = launcher.run(filedata); err != nil {
			launcher.gui.SendErrorMessage(err)
			return
		}
	}()
	if err = launcher.gui.Start(launcher.WindowTitle); err != nil {
		return err
	}
	wg.Wait()
	return err
}

// Terminate forces GUI to close
func (launcher *Launcher) Terminate() {
	if launcher.gui != nil {
		launcher.gui.Terminate()
	}
}

func (launcher *Launcher) CheckPlatform() error {
	if err := settings.EnsureJavaExecutableAvailability(); err != nil {
		return errors.Wrap(err, "java executable wasn't found")
	}
	if err := settings.EnsureJARSignerAvailability(); err != nil {
		log.Printf("%s, JAR verification will be skipped\n", err)
	}
	javaVersion, err := settings.GetJavaVersionString()
	if err != nil {
		return errors.Wrap(err, "unable to obtain java version")
	}
	log.Println(javaVersion)
	version, err := settings.GetJavaVersion()
	if err != nil {
		return err
	}
	log.Printf("Detected Java version major=%d minor=%d", version.Major, version.Minor)
	log.Printf("DisableVerification is %v", settings.IsVerificationDisabled())
	return nil
}

func (launcher *Launcher) getRelevantResources() []*Resources {
	if launcher.relevantResources == nil {
		launcher.relevantResources = launcher.jnlp.findRelevantResources()
	}
	return launcher.relevantResources
}

func (launcher *Launcher) getJars() ([]string, error) {
	return launcher.jnlp.getJars()
}

func (launcher *Launcher) getNativeLibs() ([]string, error) {
	return launcher.jnlp.getNativeLibs()
}

func (launcher *Launcher) getExtensions() ([]*Extension, error) {
	var extensions []*Extension
	codebaseURL, err := launcher.getCodebaseURL()
	if err != nil {
		return nil, err
	}
	relevantResources := launcher.getRelevantResources()
	for _, resources := range relevantResources {
		for _, extension := range resources.Extensions {
			url, err := url.Parse(extension.Href)
			if err != nil {
				continue
			}
			abs := codebaseURL.ResolveReference(url)
			extension.URL = abs.String()
			extensions = append(extensions, extension)
		}
	}
	return extensions, nil
}

func (launcher *Launcher) getCodebaseURL() (*url.URL, error) {
	if launcher.codebaseURL != nil {
		return launcher.codebaseURL, nil
	}
	codebase := launcher.jnlp.CodeBase
	codebaseURL, err := launcher_utils.ParseCodebaseURL(codebase)
	if err != nil {
		return nil, err
	}
	launcher.codebaseURL = codebaseURL
	return codebaseURL, nil
}

func (launcher *Launcher) getProperties() []Property {
	var properties []Property
	relevantResources := launcher.getRelevantResources()
	for _, resources := range relevantResources {
		for _, property := range resources.Properties {
			properties = append(properties, property)
		}
	}
	return properties
}

func (launcher *Launcher) getJVMArgs() []string {
	var jvmArgs []string
	relevantResources := launcher.getRelevantResources()
	for _, resources := range relevantResources {
		j2se := resources.getJ2SE()
		if j2se != nil && j2se.JavaVMArgs != "" {
			args := strings.Split(resources.J2SE.JavaVMArgs, " ")
			jvmArgs = append(jvmArgs, args...)
		}
	}
	return jvmArgs
}

func (launcher *Launcher) getExtensionJars() []string {
	var jars []string
	relevantResources := launcher.getRelevantResources()
	for _, resources := range relevantResources {
		for _, extension := range resources.Extensions {
			jars = append(jars, extension.Name)
		}
	}
	return jars
}

func (launcher *Launcher) command() (*exec.Cmd, error) {
	jnlp := launcher.jnlp
	jars, err := launcher.getJars()
	if err != nil {
		return nil, err
	}
	extensionJars := launcher.getExtensionJars()
	javaArgs := launcher.getJVMArgs()
	var args []string
	nativelibs, err := launcher.getNativeLibs()
	if err != nil {
		return nil, err
	}
	var nativeLibPaths []string
	for _, nativelib := range nativelibs {
		filename := path.Base(nativelib)
		filenameWithoutExt := strings.TrimSuffix(filename, path.Ext(filename))
		path := filepath.Join(launcher.resourceDir, filenameWithoutExt)
		nativeLibPaths = append(nativeLibPaths, path)
	}
	for _, jar := range jars {
		args = append(args, filepath.Join(launcher.resourceDir, path.Base(jar)))
	}
	for _, jar := range extensionJars {
		args = append(args, filepath.Join(launcher.resourceDir, path.Base(jar)))
	}
	javaArgs = append(javaArgs, "-cp", strings.Join(args, ClassPathSeparator))
	properties := launcher.getProperties()
	for _, property := range properties {
		javaArgs = append(javaArgs, fmt.Sprintf("-D%s=%s", property.Name, property.Value))
	}
	if len(nativeLibPaths) > 0 {
		javaArgs = append(javaArgs, fmt.Sprintf("-Dsettings.library.path=%s", strings.Join(nativeLibPaths, ClassPathSeparator)))
	}
	if splash := launcher.getSplashScreen(); splash != "" {
		javaArgs = append(javaArgs, fmt.Sprintf("-splash:%s", splash))
	}
	if jnlp.AppDescription != nil {
		javaArgs = append(javaArgs, jnlp.AppDescription.MainClass)
		for _, appArg := range jnlp.AppDescription.Arguments {
			javaArgs = append(javaArgs, appArg)
		}
	} else if jnlp.AppletDescription != nil {
		return nil, errors.New("found <applet-desc> tag but applets are not supported")
	} else {
		return nil, errors.New("<application-desc> tag wasn't found in JNLP file")
	}
	log.Printf("java arguments %s\n", strings.Join(javaArgs, " "))
	cmd := exec.Command(settings.Java(), javaArgs...)
	if launcher.options != nil && launcher.options.IsRunningFromBrowser {
		utils.BreakAwayFromParent(cmd)
	}
	return cmd, nil
}

func (launcher *Launcher) exec() error {
	cmd, err := launcher.command()
	if err != nil {
		return errors.Wrap(err, "unable to run java application")
	}
	launcher.cmd = cmd
	if launcher.gui.Closed() {
		return errCancelled
	}
	return cmd.Start()
}

func (launcher *Launcher) run(filedata []byte) error {
	var jnlpFile *JNLP
	var err error
	if jnlpFile, err = Decode(filedata); err != nil {
		return errors.Wrap(err, "parsing JNLP")
	}
	launcher.jnlp = jnlpFile
	launcher.filedata = filedata
	launcher.resourceDir = launcher.generateResourcesDirName(filedata)
	launcher.gui.SetTitle(launcher.jnlp.Title())
	if err := launcher.saveOriginalFile(); err != nil {
		return err
	}
	if err := launcher.estimateProgressMax(); err != nil {
		return err
	}
	if err := launcher.checkRequiredJavaVersion(); err != nil {
		return err
	}
	if err := launcher.downloadJARs(); err != nil {
		return err
	}
	if err := launcher.extractNativeLibs(); err != nil {
		return err
	}
	if err := launcher.downloadExtensions(); err != nil {
		return err
	}
	if err := launcher.downloadIcons(); err != nil {
		return err
	}
	launcher.removeOldShortcutsIfNeeded()
	if err := launcher.createShortcuts(); err != nil {
		return err
	}
	if settings.AddAppToControlPanel() {
		if err := launcher.installApp(); err != nil {
			return err
		}
	}
	if launcher.gui.Closed() {
		return errCancelled
	}
	launcher.gui.SendTextMessage("Starting application...")
	return launcher.exec()
}

func (launcher *Launcher) downloadIcons() error {
	codebaseURL, err := launcher.getCodebaseURL()
	if err != nil {
		return err
	}
	iconDir, err := launcher.createDirForResourceFiles()
	if err != nil {
		return errors.Wrapf(err, "unable to create directory for icon files")
	}
	for _, icon := range launcher.jnlp.Information.Icons {
		if launcher.gui.Closed() {
			return errCancelled
		}
		url, err := url.Parse(icon.Href)
		if err != nil {
			log.Printf("warning: unable to parse icon href %s: %v\n", icon.Href, err)
			continue
		}
		url = codebaseURL.ResolveReference(url)
		launcher.gui.SendTextMessage(fmt.Sprintf("Downloading %s", path.Base(icon.Href)))
		allowCached := true
		if _, err := download.ToFile(url.String(), iconDir, allowCached); err != nil {
			log.Printf("warning: unable to download icon %s: %v\n", icon.Href, err)
			launcher.gui.SendTextMessage(fmt.Sprintf("Warning: unable to download %s", path.Base(icon.Href)))
			continue
		}
		icon.Downloaded = true
		launcher.gui.SendTextMessage(fmt.Sprintf("Downloading %s finished", path.Base(icon.Href)))
	}
	launcher.gui.ProgressStep()
	return nil
}

func (launcher *Launcher) getSplashScreen() string {
	for _, icon := range launcher.jnlp.Information.Icons {
		if icon.Kind == "splash" {
			return filepath.Join(launcher.resourceDir, path.Base(icon.Href))
		}
	}
	return ""
}

func (launcher *Launcher) downloadJARs() error {
	jars, err := launcher.getJars()
	if err != nil {
		return err
	}
	nativeLibJars, err := launcher.getNativeLibs()
	if err != nil {
		return err
	}
	jars = append(jars, nativeLibJars...)
	jarDir, err := launcher.createDirForResourceFiles()
	if err != nil {
		return errors.Wrapf(err, "unable to create directory for jar files")
	}
	allowCached := launcher.jnlp.Information.OfflineAllowed != nil
	log.Printf("jar dir is %s\n", jarDir)
	errChan := make(chan error, len(jars))
	certChan := make(chan []byte, len(jars))
	var wg sync.WaitGroup
	wg.Add(len(jars))
	tokens := make(chan struct{}, 3)
	for _, url := range jars {
		go func(url string) {
			tokens <- struct{}{}
			defer func() { <-tokens }()
			defer wg.Done()
			if launcher.gui.Closed() {
				return
			}
			log.Printf("downloading JAR %s\n", url)
			launcher.gui.SendTextMessage(fmt.Sprintf("Downloading JAR %s\n", path.Base(url)))
			filename, err := download.ToFile(url, jarDir, allowCached)
			if err != nil {
				errChan <- err
				return
			}
			launcher.gui.ProgressStep()
			launcher.gui.SendTextMessage(fmt.Sprintf("Downloading JAR %s finished\n", path.Base(url)))
			if launcher.gui.Closed() {
				return
			}
			if !settings.IsVerificationDisabled() {
				if err := verifier.VerifyWithJARSigner(filename, false); err != nil {
					errChan <- errors.Wrapf(err, "JAR verification failed %s", filepath.Base(filename))
					return
				}
				launcher.gui.SendTextMessage(fmt.Sprintf("Checking JAR %s finished\n", path.Base(url)))
			}
			launcher.gui.ProgressStep()
			if launcher.gui.Closed() {
				return
			}
			cert, err := verifier.GetJARCertificate(filename)
			if err != nil {
				errChan <- errors.Wrapf(err, "JAR certificate error %s", filepath.Base(filename))
				return
			}
			certChan <- cert
			launcher.gui.ProgressStep()
		}(url)
	}
	wg.Wait()
	if launcher.gui.Closed() {
		return errCancelled
	}
	launcher.gui.SendTextMessage("Downloading finished")
	close(errChan)
	close(certChan)
	if err, ok := <-errChan; ok {
		return err
	}
	firstCert := <-certChan
	for cert := range certChan {
		if bytes.Equal(firstCert, cert) {
			return errors.New("all JARs have to be signed with the same certificate")
		}
	}
	launcher.cert = firstCert
	if launcher.gui.Closed() {
		return errCancelled
	}
	return nil
}

func (launcher *Launcher) downloadExtensions() error {
	launcher.gui.SendTextMessage("Downloading extensions...")
	extensions, err := launcher.getExtensions()
	allowCached := launcher.jnlp.Information.OfflineAllowed != nil
	if err != nil {
		return err
	}
	jarDir, err := launcher.createDirForResourceFiles()
	if err != nil {
		return errors.Wrapf(err, "unable to create directory for jar files")
	}
	errChan := make(chan error, len(extensions))
	var wg sync.WaitGroup
	wg.Add(len(extensions))
	tokens := make(chan struct{}, 3)
	for _, extension := range extensions {
		go func(extension *Extension) {
			tokens <- struct{}{}
			defer func() { <-tokens }()
			defer wg.Done()
			if launcher.gui.Closed() {
				return
			}
			log.Printf("downloading extension %s\n", extension.Name)
			launcher.gui.SendTextMessage(fmt.Sprintf("Downloading extension %s\n", extension.Name))
			filename, err := download.ToFile(extension.URL, jarDir, allowCached)
			if err != nil {
				errChan <- errors.Wrapf(err, "unable to download jnlp file for extension %s", extension.Name)
				return
			}
			if launcher.gui.Closed() {
				return
			}
			extensionJNLP, err := DecodeFile(filename)
			if err != nil {
				errChan <- errors.Wrapf(err, "unable to parse jnlp file for extension %s", extension.Name)
				return
			}
			jars, err := extensionJNLP.getJars()
			if err != nil {
				errChan <- errors.Wrapf(err, "unable to get JARs for extension %s", extension.Name)
				return
			}
			for _, jarURL := range jars {
				log.Printf("downloading JAR %s\n", jarURL)
				launcher.gui.SendTextMessage(fmt.Sprintf("Downloading JAR %s\n", path.Base(jarURL)))
				filename, err := download.ToFile(jarURL, jarDir, allowCached)
				if err != nil {
					errChan <- errors.Wrapf(err, "unable to download JAR for extension %s", extension.Name)
					return
				}
				launcher.gui.SendTextMessage(fmt.Sprintf("Downloading JAR %s finished\n", path.Base(jarURL)))
				if !settings.IsVerificationDisabled() {
					if err := verifier.VerifyWithJARSigner(filename, false); err != nil {
						errChan <- errors.Wrapf(err, "JAR verification failed %s", filepath.Base(filename))
						return
					}
				}
				cert, err := verifier.GetJARCertificate(filename)
				if err != nil {
					errChan <- errors.Wrapf(err, "JAR certificate error %s", filepath.Base(filename))
					return
				}
				if bytes.Equal(launcher.cert, cert) {
					errChan <- errors.New("all JARs have to be signed with the same certificate")
					return
				}
				launcher.gui.SendTextMessage(fmt.Sprintf("Checking JAR %s finished\n", path.Base(jarURL)))
				if launcher.gui.Closed() {
					return
				}
			}
			launcher.gui.ProgressStep()
			launcher.gui.SendTextMessage(fmt.Sprintf("Downloading extension %s finished\n", extension.Name))
		}(extension)
	}
	wg.Wait()
	if launcher.gui.Closed() {
		return errCancelled
	}
	close(errChan)
	if err, ok := <-errChan; ok {
		return err
	}
	if launcher.gui.Closed() {
		return errCancelled
	}
	return nil
}

func (launcher *Launcher) extractNativeLibs() error {
	nativeLibJars, err := launcher.getNativeLibs()
	if err != nil {
		return err
	}
	jarDir := launcher.resourceDir
	for _, url := range nativeLibJars {
		if launcher.gui.Closed() {
			return errCancelled
		}
		log.Printf("extracting Nativelib %s\n", path.Base(url))
		filename := path.Base(url)
		filenameWithoutExt := strings.TrimSuffix(filename, path.Ext(filename))
		dir := filepath.Join(jarDir, filenameWithoutExt)
		zipFilename := filepath.Join(launcher.resourceDir, path.Base(url))
		launcher.gui.SendTextMessage(fmt.Sprintf("Extracting Nativelib %s\n", path.Base(url)))
		if err := launcher_utils.Extract(zipFilename, dir); err != nil {
			return errors.Wrapf(err, "extracting nativelib %s", path.Base(url))
		}
	}
	if launcher.gui.Closed() {
		return errCancelled
	}
	return nil
}

func (launcher *Launcher) estimateProgressMax() error {
	jars, err := launcher.getJars()
	if err != nil {
		return err
	}
	nativeLibJars, err := launcher.getNativeLibs()
	if err != nil {
		return err
	}
	extensionJars := launcher.getExtensionJars()
	progressMax := 3*(len(jars)+len(nativeLibJars)) + len(extensionJars) + 1
	launcher.gui.SetProgressMax(progressMax)
	return nil
}

func (launcher *Launcher) createShortcuts() error {
	info := launcher.jnlp.Information
	title := info.Title
	iconSrc := launcher.findShortcutIcon()
	description := launcher.getShortcutDescription()
	arguments := launcher.getShortcutArguments()
	if info.Desktop != nil || (info.Shortcut != nil && info.Shortcut.Desktop != nil) {
		launcher.gui.SendTextMessage("Creating Desktop shortcut")
		if err := utils.CreateDesktopShortcut(os.Args[0], title, description, iconSrc, arguments...); err != nil {
			return err
		}
	}
	if info.Shortcut != nil && info.Shortcut.Menu != nil {
		submenu := info.Shortcut.Menu.SubMenu
		launcher.gui.SendTextMessage("Creating Start Menu shortcut")
		if err := utils.CreateStartMenuShortcut(os.Args[0], submenu, title, description, iconSrc, arguments...); err != nil {
			return err
		}
	}
	return nil
}

func (launcher *Launcher) removeOldShortcutsIfNeeded() {
	jnlpOld := launcher.jnlpOld
	if jnlpOld == nil {
		return
	}
	launcher.removeShortcuts(jnlpOld)
}

func (launcher *Launcher) removeShortcuts(jnlp *JNLP) {
	// important: the method may run without GUI
	info := jnlp.Information
	title := info.Title
	if info.Desktop != nil || (info.Shortcut != nil && info.Shortcut.Desktop != nil) {
		log.Printf("removing old desktop shortcut: %s", title)
		if err := utils.RemoveDesktopShortcut(title); err != nil {
			log.Printf("warning: error while removing old desktop shortcut: %v", err)
		}
	}
	if info.Shortcut != nil && info.Shortcut.Menu != nil {
		submenu := info.Shortcut.Menu.SubMenu
		log.Printf("removing old start menu folder: %s", submenu)
		if err := utils.RemoveStartMenuFolder(submenu); err != nil {
			log.Printf("warning: error while removing old start menu folder: %v", err)
		}
	}
}

func (launcher *Launcher) findShortcutIcon() string {
	iconDir := launcher.resourceDir
	for _, icon := range launcher.jnlp.Information.Icons {
		if icon.Kind != "" && icon.Kind != "default" && icon.Kind != "shortcut" {
			continue
		}
		if !icon.Downloaded {
			continue
		}
		if path.Ext(icon.Href) != ".ico" {
			continue
		}
		return filepath.Join(iconDir, path.Base(icon.Href))

	}
	return ""
}

func (launcher *Launcher) getShortcutDescription() string {
	info := launcher.jnlp.Information
	descriptions := info.Descriptions
	descriptionMap := make(map[string]string)
	for _, desc := range descriptions {
		if desc.Text != "" {
			descriptionMap[desc.Kind] = desc.Text
		}
	}
	if desc, ok := descriptionMap["tooltip"]; ok {
		return desc
	}
	if desc, ok := descriptionMap["short"]; ok {
		return desc
	}
	if desc, ok := descriptionMap[""]; ok {
		return desc
	}
	return ""
}

func (launcher *Launcher) getShortcutArguments() []string {
	var arguments []string
	if launcher.options != nil && launcher.options.JavaDir != "" {
		arguments = append(arguments, "-javadir")
		arguments = append(arguments, launcher.options.JavaDir)
	}
	if launcher.options != nil && launcher.options.ShowConsole {
		arguments = append(arguments, "-showconsole")
	}
	arguments = append(arguments, launcher.getOriginalFilePath())
	return arguments
}

func (launcher *Launcher) createDirForResourceFiles() (string, error) {
	resourceDir := launcher.resourceDir
	if _, err := os.Stat(resourceDir); os.IsNotExist(err) {
		err = os.MkdirAll(resourceDir, 0755)
		if err != nil {
			return "", errors.Wrapf(err, "unable to create directory for resource files '%s'", resourceDir)
		}
	}
	return resourceDir, nil
}

func (launcher *Launcher) checkForUpdate(filedata []byte) ([]byte, error) {
	var err error
	var jnlpFile *JNLP
	if jnlpFile, err = Decode(filedata); err != nil {
		return nil, errors.Wrap(err, "parsing JNLP")
	}
	if jnlpFile.Href == "" {
		log.Printf("warning: unable to check jnlp file for update because <jnlp> tag doesn't have 'href' attribute or the attribute is empty")
		return filedata, nil
	}
	var codeBaseURL *url.URL
	codeBaseURL, err = launcher_utils.ParseCodebaseURL(jnlpFile.CodeBase)
	if err != nil {
		return nil, err
	}
	var hrefURL *url.URL
	hrefURL, err = url.Parse(jnlpFile.Href)
	if err != nil {
		log.Printf("warning: unable to check jnlp file for update because 'href' attribute on <jnlp> tag is invalid: %v", err)
		return filedata, nil
	}
	jnlpURL := codeBaseURL.ResolveReference(hrefURL)
	var newFileData []byte
	if newFileData, err = download.ToMemory(jnlpURL.String()); err != nil {
		log.Printf("warning: unable to check jnlp file for update because %v", err)
		return filedata, nil
	}
	if bytes.Compare(filedata, newFileData) == 0 {
		log.Printf("jnlp file hasn't been changed")
		return filedata, nil
	}
	log.Printf("jnlp file has been changed")
	if _, err = Decode(newFileData); err != nil {
		log.Printf("warning: unable to parse new jnlp file because %v, existing copy will be used", err)
		return filedata, nil
	}
	log.Printf("jnlp file updated successfully")
	launcher.jnlpOld = jnlpFile
	return newFileData, nil
}

func (launcher *Launcher) generateResourcesDirName(filedata []byte) string {
	return launcher_utils.GenerateResourcesDirName(launcher.WorkDir, filedata)
}

func (launcher *Launcher) getOriginalFilePath() string {
	return filepath.Join(launcher.resourceDir, "original.jnlp")
}

func (launcher *Launcher) installApp() error {
	info := launcher.jnlp.Information
	uninstallString := os.Args[0] + " -uninstall -gui " + launcher.getOriginalFilePath()
	url := ""
	if info.Homepage != nil {
		url = info.Homepage.Href
	}
	app := &utils.AppInfo{
		Title: info.Title,
		UninstallString: uninstallString,
		Icon: launcher.findShortcutIcon(),
		Version: info.Version,
		URL: url,
		Publisher: info.Vendor,
	}
	log.Printf("adding app into Control Panel")
	if err := utils.InstallApp(app); err != nil {
		return errors.Wrap(err, "unable to install app into control panel")
	}
	return nil
}

func (launcher *Launcher) uninstallApp(jnlp *JNLP) error {
	title := jnlp.Title()
	if title != "" {
		if err := utils.UninstallApp(title); err != nil {
			return errors.Wrap(err, "unable to uninstall app from control panel")
		}
	}
	return nil
}

func (launcher *Launcher) saveOriginalFile() error {
	_, err := launcher.createDirForResourceFiles()
	if err != nil {
		return errors.Wrapf(err, "unable to create directory for resource files")
	}
	originalFilePath := launcher.getOriginalFilePath()
	if err := ioutil.WriteFile(originalFilePath, launcher.filedata, 0644); err != nil {
		return errors.Wrap(err, "unable to save original jnlp file")
	}
	return nil
}

func (launcher *Launcher) checkRequiredJavaVersion() error {
	relevantResources := launcher.getRelevantResources()
	for _, resources := range relevantResources {
		j2se := resources.getJ2SE()
		if j2se != nil {
			requiredVersion, err  := settings.ParseJavaVersion(j2se.Version)
			if err != nil {
				return errors.Wrapf(err, `unable to parse version="%s" in jnlp file`, j2se.Version)
			}
			if !settings.CurrentJavaVersionMatches(requiredVersion) {
				currentVersion, _ := settings.GetJavaVersion()
				return errors.Errorf("This JNLP file requires Java version %s, found %s", j2se.Version, currentVersion.String)
			}
		}
	}
	return nil
}

func (launcher *Launcher) normalizeURL(url string) string {
	normalizedURL := url
	if strings.HasPrefix(url, "jnlp://") {
		normalizedURL = strings.Replace(url, "jnlp://", "http://", 1)
	} else if strings.HasPrefix(url, "jnlps://") {
		normalizedURL = strings.Replace(url, "jnlps://", "https://", 1)
	}
	return normalizedURL
}

func init() {
	jnlpLauncher := NewLauncher()
	launcher.RegisterProtocol("jnlp", jnlpLauncher)
	launcher.RegisterProtocol("jnlps", jnlpLauncher)
	launcher.RegisterExtension("jnlp", jnlpLauncher)
}
