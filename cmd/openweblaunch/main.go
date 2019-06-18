package main

import (

	_ "github.com/rocketsoftware/open-web-launch/launcher/jnlp"
	"github.com/rocketsoftware/open-web-launch/bootstrap"
)

//go:generate goversioninfo -o openweblaunch.syso

const (
	productName  = "openweblaunch"
	productTitle = "Open Web Launch"
)

var (
	productVersion = "Dummy version number"
)

func main() {
	bootstrap.Run(productName, productTitle, productVersion)
}
