package main

import (
	"github.com/rocketsoftware/open-web-launch/bootstrap"
	_ "github.com/rocketsoftware/open-web-launch/launcher/jnlp"
)

//go:generate goversioninfo -o openweblaunch.syso

var (
	productName    = "openweblaunch"
	productTitle   = "Open Web Launch"
	productVersion = "Dummy version number"
)

func main() {
	bootstrap.Run(productName, productTitle, productVersion)
}
