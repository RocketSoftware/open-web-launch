package bootstrap

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
	"github.com/rocketsoftware/open-web-launch/java"
	"github.com/rocketsoftware/open-web-launch/launcher"
	"github.com/rocketsoftware/open-web-launch/messaging"
	"github.com/rocketsoftware/open-web-launch/utils"
)

var (
	javaDir string
	showConsole bool
	uninstall bool
)

func Run(productName, productTitle, productVersion string) {
	productWorkDir := filepath.Join(os.TempDir(), productName)
	productLogFile := filepath.Join(productWorkDir, productName+".log")
	fmt.Fprintf(os.Stderr, "%s %s\n", productTitle, productVersion)
	if err := utils.CreateProductWorkDir(productWorkDir); err != nil {
		log.Fatal(err)
	}
	logFile, err := utils.OpenOrCreateProductLogFile(productLogFile)
	if err != nil {
		log.Fatal(err)
	}
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(logFile)
	log.Printf("starting %s %s with arguments %v\n", productTitle, productVersion, os.Args)
	log.Printf("current platform is OS=%q Architecture=%q\n", runtime.GOOS, runtime.GOARCH)
	flag.BoolVar(&showConsole, "showconsole", false, "show Java console")
	flag.StringVar(&javaDir, "javadir", "", "Java folder that should be used for starting a Java Web Start application")
	flag.BoolVar(&uninstall, "uninstall", false, "uninstall a specific Java Web Start application")
	flag.Parse()
	nargs := flag.NArg()
	nflags := flag.NFlag()
	if nargs == 1 && nflags == 0 {
		filenameOrURL := flag.Arg(0)
		handleURLOrFilename(filenameOrURL, nil, productWorkDir, productTitle)
	} else if nargs == 1 && uninstall {
		filenameOrURL := flag.Arg(0)
		handleUninstallCommand(filenameOrURL, productWorkDir, productTitle)
	} else if nargs == 1 {
		filenameOrURL := flag.Arg(0)
		options := &launcher.Options{}
		if isFlagSet("javadir") {
			var err error
			if javaDir, err = java.UseJavaDir(javaDir); err != nil {
				log.Fatal(err)
			}
			options.JavaDir = javaDir
		}
		if isFlagSet("showconsole") {
			java.ShowConsole()
			options.ShowConsole = true
		}
		handleURLOrFilename(filenameOrURL, options, productWorkDir, productTitle)
	} else {
		isRunningFromBrowser := len(os.Args) > 2
		options := &launcher.Options{IsRunningFromBrowser: isRunningFromBrowser}
		log.Printf("running from browser: %v\n", isRunningFromBrowser)
		listenForMessage(options, productWorkDir, productTitle)
	}
}

func handleURLOrFilename(filenameOrURL string, options *launcher.Options, productWorkDir string, productTitle string) {
	myLauncher, byURL, err := launcher.FindLauncherForURLOrFilename(filenameOrURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := myLauncher.CheckPlatform(); err != nil {
		log.Fatal(err)
	}
	myLauncher.SetWorkDir(productWorkDir)
	myLauncher.SetWindowTitle(productTitle)
	myLauncher.SetOptions(options)
	defer myLauncher.Wait()
	if byURL {
		if err := myLauncher.RunByURL(filenameOrURL); err != nil {
			log.Println(err)
			return
		}
	} else {
		if err := myLauncher.RunByFilename(filenameOrURL); err != nil {
			log.Println(err)
			return
		}
	}
}

func listenForMessage(options *launcher.Options, productWorkDir string, productTitle string) {
	message, err := messaging.GetMessage(os.Stdin)
	if err != nil {
		if errors.Cause(err) != io.EOF {
			log.Fatal(err)
		}
		log.Println("exit because stdin has been closed")
		return
	}
	if message.Status != "" {
		response := fmt.Sprintf(`{"status": "installed"}`)
		if err := messaging.SendMessage(os.Stdout, response); err != nil {
			log.Fatal(err)
		}
		return
	}
	myLauncher, err := launcher.FindLauncherForURL(message.URL)
	if err != nil {
		log.Fatal(err)
	}
	if err := myLauncher.CheckPlatform(); err != nil {
		log.Fatal(err)
	}
	myLauncher.SetWorkDir(productWorkDir)
	myLauncher.SetWindowTitle(productTitle)
	myLauncher.SetOptions(options)
	defer myLauncher.Wait()
	if err := myLauncher.RunByURL(message.URL); err != nil {
		stringError := fmt.Sprintf("%v", err)
		jsonError, _ := json.Marshal(stringError)
		response := fmt.Sprintf(`{"status": %s}`, string(jsonError))
		log.Println(response)
		if err := messaging.SendMessage(os.Stdout, response); err != nil {
			log.Fatal(err)
		}
		return
	}
	response := fmt.Sprintf(`{"status": "ok"}`)
	if err := messaging.SendMessage(os.Stdout, response); err != nil {
		log.Fatal(err)
	}
}

func handleUninstallCommand(filenameOrURL string, productWorkDir string, productTitle string) {
	myLauncher, byURL, err := launcher.FindLauncherForURLOrFilename(filenameOrURL)
	if err != nil {
		log.Fatal(err)
	}
	myLauncher.SetWorkDir(productWorkDir)
	myLauncher.SetWindowTitle(productTitle)
	if byURL {
		if err := myLauncher.UninstallByURL(filenameOrURL); err != nil {
			log.Println(err)
			return
		}
	} else {
		if err := myLauncher.UninstallByFilename(filenameOrURL); err != nil {
			log.Println(err)
			return
		}
	}
}


func isFlagSet(flagName string) bool {
	found := false
    flag.Visit(func(f *flag.Flag) {
        if f.Name == flagName {
            found = true
        }
    })
    return found
}