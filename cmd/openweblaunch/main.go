package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"

	"github.com/rocketsoftware/open-web-launch/java"
	"github.com/rocketsoftware/open-web-launch/launcher"
	_ "github.com/rocketsoftware/open-web-launch/launcher/jnlp"
	"github.com/rocketsoftware/open-web-launch/messaging"
	"github.com/pkg/errors"
)

func main() {
	fmt.Fprintf(os.Stderr, "%s %s\n", productTitle, productVersion)
	if err := createProductWorkDir(); err != nil {
		log.Fatal(err)
	}
	logFile, err := openOrCreateProductLogFile()
	if err != nil {
		log.Fatal(err)
	}
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(logFile)
	log.Printf("starting %s %s with arguments %v\n", productTitle, productVersion, os.Args)
	log.Printf("current platform is OS=%q Architecture=%q\n", runtime.GOOS, runtime.GOARCH)
	if len(os.Args) == 2 {
		filenameOrURL := os.Args[1]
		handleURLOrFilename(filenameOrURL, nil)
	} else if len(os.Args) == 3 && os.Args[1] == "-uninstall" {
		handleUninstallCommand()
	} else if len(os.Args) == 4 && os.Args[1] == "-javadir" {
		javaDir := os.Args[2]
		filenameOrURL := os.Args[3]
		var err error
		if javaDir, err = java.UseJavaDir(javaDir); err != nil {
			log.Fatal(err)
		}
		options := &launcher.Options{JavaDir: javaDir}
		handleURLOrFilename(filenameOrURL, options)
	} else {
		isRunningFromBrowser := len(os.Args) > 2
		options := &launcher.Options{IsRunningFromBrowser: isRunningFromBrowser}
		log.Printf("running from browser: %v\n", isRunningFromBrowser)
		listenForMessage(options)
	}
}

func handleURLOrFilename(filenameOrURL string, options *launcher.Options) {
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

func listenForMessage(options *launcher.Options) {
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

func handleUninstallCommand() {
	_ = os.Args[1] // -uninstall
	filenameOrURL := os.Args[2]
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
