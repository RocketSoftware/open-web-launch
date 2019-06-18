package main

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

//go:generate goversioninfo -o openweblaunch.syso

const (
	productName  = "openweblaunch"
	productTitle = "Open Web Launch"
)

var (
	productVersion = "Dummy version number"
	productWorkDir = filepath.Join(os.TempDir(), productName)
	productLogFile = filepath.Join(productWorkDir, productName+".log")
)

func createProductWorkDir() error {
	if _, err := os.Stat(productWorkDir); os.IsNotExist(err) {
		err = os.MkdirAll(productWorkDir, 0755)
		if err != nil {
			return errors.Wrapf(err, "unable to create product working directory %q", productWorkDir)
		}
	}
	return nil
}

func openOrCreateProductLogFile() (*os.File, error) {
	return os.OpenFile(productLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)

}
